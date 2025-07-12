-- Enhanced Payment System Migration

-- Drop existing payments table if it exists
DROP TABLE IF EXISTS payments CASCADE;

-- Create enhanced payments table
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    plan_id INTEGER REFERENCES subscription_plans(id) ON DELETE SET NULL,
    amount INTEGER NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    payment_method VARCHAR(50) NOT NULL,
    payment_provider VARCHAR(50) NOT NULL,
    transaction_id VARCHAR(255) UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled', 'refunded')),
    description TEXT,
    payment_data JSONB,
    webhook_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payments_transaction_id ON payments(transaction_id);
CREATE INDEX idx_payments_provider ON payments(payment_provider);

-- Create payment providers table
CREATE TABLE payment_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    configuration JSONB,
    webhook_secret TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default payment providers
INSERT INTO payment_providers (name, display_name, is_active, configuration) VALUES
('stripe', 'Credit Card (Stripe)', TRUE, '{"supports_webhooks": true, "currencies": ["USD", "EUR", "GBP"]}'),
('yoomoney', 'ЮMoney', TRUE, '{"supports_webhooks": true, "currencies": ["RUB"]}'),
('paypal', 'PayPal', TRUE, '{"supports_webhooks": true, "currencies": ["USD", "EUR", "GBP"]}'),
('crypto', 'Cryptocurrency', TRUE, '{"supports_webhooks": true, "currencies": ["BTC", "ETH", "USDT"]}'),
('telegram', 'Telegram Payments', TRUE, '{"supports_webhooks": true, "currencies": ["USD", "EUR", "RUB"]}');

-- Create subscription activations table
CREATE TABLE subscription_activations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    payment_id INTEGER REFERENCES payments(id) ON DELETE CASCADE,
    plan_id INTEGER REFERENCES subscription_plans(id) ON DELETE SET NULL,
    activated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for subscription activations
CREATE INDEX idx_subscription_activations_user_id ON subscription_activations(user_id);
CREATE INDEX idx_subscription_activations_expires_at ON subscription_activations(expires_at);
CREATE INDEX idx_subscription_activations_is_active ON subscription_activations(is_active);

-- Create payment webhooks log table
CREATE TABLE payment_webhook_logs (
    id SERIAL PRIMARY KEY,
    payment_provider VARCHAR(50) NOT NULL,
    webhook_id VARCHAR(255),
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    headers JSONB,
    processed BOOLEAN DEFAULT FALSE,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for webhook logs
CREATE INDEX idx_payment_webhook_logs_provider ON payment_webhook_logs(payment_provider);
CREATE INDEX idx_payment_webhook_logs_created_at ON payment_webhook_logs(created_at);
CREATE INDEX idx_payment_webhook_logs_processed ON payment_webhook_logs(processed);

-- Create payment statistics view
CREATE OR REPLACE VIEW payment_statistics AS
SELECT 
    DATE(created_at) as date,
    payment_provider,
    COUNT(*) as total_payments,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as successful_payments,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_payments,
    SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END) as total_revenue,
    AVG(CASE WHEN status = 'completed' THEN amount END) as avg_payment_amount,
    ROUND(
        (COUNT(CASE WHEN status = 'completed' THEN 1 END) * 100.0 / NULLIF(COUNT(*), 0)), 2
    ) as success_rate
FROM payments
WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY DATE(created_at), payment_provider
ORDER BY date DESC, payment_provider;

-- Create function to update payment updated_at timestamp
CREATE OR REPLACE FUNCTION update_payment_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for payments table
CREATE TRIGGER trigger_update_payment_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_payment_updated_at();

-- Create function to automatically create subscription activation
CREATE OR REPLACE FUNCTION create_subscription_activation()
RETURNS TRIGGER AS $$
BEGIN
    -- Only create activation when payment is completed
    IF NEW.status = 'completed' AND OLD.status != 'completed' THEN
        INSERT INTO subscription_activations (user_id, payment_id, plan_id, expires_at)
        SELECT 
            NEW.user_id, 
            NEW.id, 
            NEW.plan_id,
            CURRENT_TIMESTAMP + INTERVAL '1 day' * sp.duration_days
        FROM subscription_plans sp
        WHERE sp.id = NEW.plan_id;
        
        -- Update user's subscription info
        UPDATE users 
        SET 
            plan_name = (SELECT name FROM subscription_plans WHERE id = NEW.plan_id),
            plan_expires_at = CURRENT_TIMESTAMP + INTERVAL '1 day' * (SELECT duration_days FROM subscription_plans WHERE id = NEW.plan_id),
            total_spent = total_spent + NEW.amount
        WHERE id = NEW.user_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for automatic subscription activation
CREATE TRIGGER trigger_create_subscription_activation
    AFTER UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION create_subscription_activation();

-- Create function to get payment analytics
CREATE OR REPLACE FUNCTION get_payment_analytics(days_back INTEGER DEFAULT 30)
RETURNS TABLE (
    total_revenue BIGINT,
    total_payments INTEGER,
    successful_payments INTEGER,
    failed_payments INTEGER,
    success_rate NUMERIC,
    avg_payment_amount NUMERIC,
    top_payment_method TEXT,
    daily_revenue JSONB
) AS $$
BEGIN
    RETURN QUERY
    WITH payment_stats AS (
        SELECT 
            COALESCE(SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END), 0) as total_rev,
            COUNT(*) as total_pay,
            COUNT(CASE WHEN status = 'completed' THEN 1 END) as success_pay,
            COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_pay,
            AVG(CASE WHEN status = 'completed' THEN amount END) as avg_amount
        FROM payments
        WHERE created_at >= CURRENT_DATE - INTERVAL '1 day' * days_back
    ),
    top_method AS (
        SELECT payment_method as method
        FROM payments
        WHERE created_at >= CURRENT_DATE - INTERVAL '1 day' * days_back
        AND status = 'completed'
        GROUP BY payment_method
        ORDER BY COUNT(*) DESC
        LIMIT 1
    ),
    daily_stats AS (
        SELECT json_agg(
            json_build_object(
                'date', date,
                'revenue', revenue,
                'payments', payments
            ) ORDER BY date
        ) as daily_data
        FROM (
            SELECT 
                DATE(created_at) as date,
                COALESCE(SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END), 0) as revenue,
                COUNT(*) as payments
            FROM payments
            WHERE created_at >= CURRENT_DATE - INTERVAL '1 day' * days_back
            GROUP BY DATE(created_at)
            ORDER BY DATE(created_at)
        ) daily_grouped
    )
    SELECT 
        ps.total_rev,
        ps.total_pay,
        ps.success_pay,
        ps.failed_pay,
        ROUND(
            (ps.success_pay * 100.0 / NULLIF(ps.total_pay, 0)), 2
        ),
        ROUND(ps.avg_amount, 2),
        COALESCE(tm.method, 'N/A'),
        COALESCE(ds.daily_data, '[]'::jsonb)
    FROM payment_stats ps
    CROSS JOIN top_method tm
    CROSS JOIN daily_stats ds;
END;
$$ LANGUAGE plpgsql;

-- Create indexes for better query performance
CREATE INDEX idx_users_plan_expires_at ON users(plan_expires_at) WHERE plan_expires_at IS NOT NULL;
CREATE INDEX idx_users_total_spent ON users(total_spent);

-- Update existing data if needed
UPDATE users SET total_spent = 0 WHERE total_spent IS NULL;
UPDATE users SET plan_name = 'Free' WHERE plan_name IS NULL OR plan_name = '';

-- Create payment summary view for dashboard
CREATE OR REPLACE VIEW payment_summary AS
SELECT 
    u.id as user_id,
    u.username,
    u.first_name,
    u.plan_name,
    u.plan_expires_at,
    u.total_spent,
    COUNT(p.id) as total_payments,
    MAX(p.created_at) as last_payment_date,
    COALESCE(SUM(CASE WHEN p.status = 'completed' THEN p.amount ELSE 0 END), 0) as total_revenue
FROM users u
LEFT JOIN payments p ON u.id = p.user_id
GROUP BY u.id, u.username, u.first_name, u.plan_name, u.plan_expires_at, u.total_spent;

-- Grant necessary permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON payments TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON payment_providers TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON subscription_activations TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON payment_webhook_logs TO PUBLIC;
GRANT SELECT ON payment_statistics TO PUBLIC;
GRANT SELECT ON payment_summary TO PUBLIC;
GRANT USAGE ON SEQUENCE payments_id_seq TO PUBLIC;
GRANT USAGE ON SEQUENCE payment_providers_id_seq TO PUBLIC;
GRANT USAGE ON SEQUENCE subscription_activations_id_seq TO PUBLIC;
GRANT USAGE ON SEQUENCE payment_webhook_logs_id_seq TO PUBLIC;