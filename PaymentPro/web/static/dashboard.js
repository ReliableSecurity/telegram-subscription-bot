// Dashboard JavaScript
let currentToken = localStorage.getItem('authToken') || new URLSearchParams(window.location.search).get('token');
let currentUserId = null;

// Save token if found in URL
if (currentToken && !localStorage.getItem('authToken')) {
    localStorage.setItem('authToken', currentToken);
}

// Set auth token for API requests
function setAuthHeaders() {
    return {
        'Authorization': currentToken,
        'Content-Type': 'application/json'
    };
}

// Enhanced fetch with token in URL as fallback
function fetchWithToken(url, options = {}) {
    // Add token to URL as query parameter as fallback
    const tokenUrl = url.includes('?') ? `${url}&token=${currentToken}` : `${url}?token=${currentToken}`;
    
    // Also add to headers
    const headers = {
        'Authorization': currentToken,
        'Content-Type': 'application/json',
        ...options.headers
    };
    
    return fetch(tokenUrl, {
        ...options,
        headers
    });
}

// Navigation
document.addEventListener('DOMContentLoaded', function() {
    console.log('DOM loaded, initializing...');
    
    // Check for token in URL first
    const urlToken = new URLSearchParams(window.location.search).get('token');
    if (urlToken) {
        console.log('Token found in URL, saving to localStorage');
        currentToken = urlToken;
        localStorage.setItem('authToken', urlToken);
    }
    
    console.log('Checking auth token:', currentToken);
    
    // Check if user is logged in
    if (!currentToken) {
        console.log('No token found, redirecting to login');
        window.location.href = '/';
        return;
    }
    
    console.log('Auth check passed, loading data...');
    
    // Load initial data
    refreshData();
    
    // Setup navigation
    setupNavigation();
    
    // Setup refresh timer
    setInterval(refreshData, 30000); // Refresh every 30 seconds
});

function setupNavigation() {
    const navLinks = document.querySelectorAll('.nav-link');
    const sections = document.querySelectorAll('.section');
    
    navLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            
            // Remove active class from all links and sections
            navLinks.forEach(l => l.classList.remove('active'));
            sections.forEach(s => s.classList.remove('active'));
            
            // Add active class to clicked link
            this.classList.add('active');
            
            // Show corresponding section
            const sectionId = this.getAttribute('data-section') + '-section';
            document.getElementById(sectionId).classList.add('active');
            
            // Update page title
            const pageTitle = this.textContent.trim();
            document.getElementById('page-title').textContent = pageTitle;
            
            // Load section data
            loadSectionData(this.getAttribute('data-section'));
        });
    });
}

function loadSectionData(section) {
    switch(section) {
        case 'overview':
            loadStats();
            break;
        case 'users':
            loadUsers();
            break;
        case 'payments':
            loadPayments();
            break;
        case 'plans':
            loadPlans();
            break;
        case 'analytics':
            loadAnalytics();
            break;
        case 'system':
            loadSystemInfo();
            break;
    }
}

async function refreshData() {
    await loadStats();
    const activeSection = document.querySelector('.nav-link.active');
    if (activeSection) {
        loadSectionData(activeSection.getAttribute('data-section'));
    }
}

async function loadStats() {
    try {
        console.log('Loading stats with token:', currentToken);
        const response = await fetchWithToken('/api/stats');
        
        console.log('Stats response status:', response.status);
        console.log('Stats response ok:', response.ok);
        
        if (!response.ok) {
            if (response.status === 401) {
                console.log('Authentication failed, redirecting to login');
                localStorage.removeItem('authToken');
                window.location.href = '/?error=auth';
                return;
            }
            throw new Error('Failed to fetch stats');
        }
        
        const stats = await response.json();
        console.log('Stats loaded:', stats);
        updateStatsDisplay(stats);
    } catch (error) {
        console.error('Error loading stats:', error);
        showNotification('Ошибка загрузки статистики', 'error');
    }
}

function updateStatsDisplay(stats) {
    document.getElementById('total-users').textContent = stats.total_users || 0;
    document.getElementById('today-users').textContent = `+${stats.today_users || 0} today`;
    document.getElementById('active-subscriptions').textContent = stats.active_subscriptions || 0;
    document.getElementById('expiring-soon').textContent = `${stats.expiring_soon || 0} expiring soon`;
    document.getElementById('total-payments').textContent = stats.total_payments || 0;
    document.getElementById('today-payments').textContent = `+${stats.today_payments || 0} today`;
    document.getElementById('total-revenue').textContent = `$${(stats.total_revenue || 0).toFixed(2)}`;
    document.getElementById('today-revenue').textContent = `+$${(stats.today_revenue || 0).toFixed(2)} today`;
}

async function loadUsers() {
    try {
        const response = await fetchWithToken('/api/users');
        
        if (!response.ok) {
            if (response.status === 401) {
                localStorage.removeItem('authToken');
                window.location.href = '/?error=auth';
                return;
            }
            throw new Error('Failed to fetch users');
        }
        
        const users = await response.json();
        updateUsersTable(users);
    } catch (error) {
        console.error('Error loading users:', error);
        showNotification('Ошибка загрузки пользователей', 'error');
    }
}

function updateUsersTable(users) {
    const tbody = document.querySelector('#users-table tbody');
    tbody.innerHTML = '';
    
    users.forEach(user => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${user.id}</td>
            <td>${user.username || 'N/A'}</td>
            <td>${user.first_name || 'N/A'}</td>
            <td>${user.plan_name || 'Free'}</td>
            <td>${user.plan_expires ? new Date(user.plan_expires).toLocaleDateString() : 'Never'}</td>
            <td>$${(user.total_spent || 0).toFixed(2)}</td>
            <td>
                <button class="btn btn-sm btn-primary" onclick="showUserActions(${user.id})">
                    <i class="fas fa-cog"></i>
                </button>
            </td>
        `;
        tbody.appendChild(row);
    });
}

async function loadPayments() {
    try {
        const response = await fetchWithToken('/api/payments');
        
        if (!response.ok) {
            if (response.status === 401) {
                localStorage.removeItem('authToken');
                window.location.href = '/?error=auth';
                return;
            }
            throw new Error('Failed to fetch payments');
        }
        
        const payments = await response.json();
        updatePaymentsTable(payments);
    } catch (error) {
        console.error('Error loading payments:', error);
        showNotification('Ошибка загрузки платежей', 'error');
    }
}

function updatePaymentsTable(payments) {
    const tbody = document.querySelector('#payments-table tbody');
    tbody.innerHTML = '';
    
    payments.forEach(payment => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${payment.id}</td>
            <td>${payment.user_name || 'N/A'}</td>
            <td>$${(payment.amount || 0).toFixed(2)}</td>
            <td>${payment.currency || 'USD'}</td>
            <td>${payment.payment_method || 'N/A'}</td>
            <td><span class="status ${payment.status}">${payment.status}</span></td>
            <td>${new Date(payment.created_at).toLocaleDateString()}</td>
        `;
        tbody.appendChild(row);
    });
}

async function loadPlans() {
    try {
        const response = await fetchWithToken('/api/plans');
        
        if (!response.ok) {
            if (response.status === 401) {
                localStorage.removeItem('authToken');
                window.location.href = '/?error=auth';
                return;
            }
            throw new Error('Failed to fetch plans');
        }
        
        const plans = await response.json();
        updatePlansTable(plans);
    } catch (error) {
        console.error('Error loading plans:', error);
        showNotification('Ошибка загрузки планов', 'error');
    }
}

function updatePlansTable(plans) {
    const tbody = document.querySelector('#plans-table tbody');
    tbody.innerHTML = '';
    
    plans.forEach(plan => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${plan.id}</td>
            <td>${plan.name}</td>
            <td>$${(plan.price_cents / 100).toFixed(2)}</td>
            <td>${plan.duration_days} days</td>
            <td>${plan.max_groups}</td>
            <td>${plan.currency}</td>
            <td>
                <button class="btn btn-sm btn-primary" onclick="editPlan(${plan.id})">
                    <i class="fas fa-edit"></i>
                </button>
                <button class="btn btn-sm btn-danger" onclick="deletePlan(${plan.id})">
                    <i class="fas fa-trash"></i>
                </button>
            </td>
        `;
        tbody.appendChild(row);
    });
}

async function loadAnalytics() {
    try {
        const [revenueResponse, usersResponse] = await Promise.all([
            fetch('/api/revenue-chart', { headers: setAuthHeaders() }),
            fetch('/api/users-chart', { headers: setAuthHeaders() })
        ]);
        
        if (!revenueResponse.ok || !usersResponse.ok) {
            throw new Error('Failed to fetch analytics');
        }
        
        const revenueData = await revenueResponse.json();
        const usersData = await usersResponse.json();
        
        updateCharts(revenueData, usersData);
    } catch (error) {
        console.error('Error loading analytics:', error);
        showNotification('Error loading analytics', 'error');
    }
}

function updateCharts(revenueData, usersData) {
    // Revenue Chart
    const revenueCtx = document.getElementById('revenue-chart').getContext('2d');
    new Chart(revenueCtx, {
        type: 'line',
        data: {
            labels: revenueData.labels,
            datasets: [{
                label: 'Revenue ($)',
                data: revenueData.data,
                borderColor: '#667eea',
                backgroundColor: 'rgba(102, 126, 234, 0.1)',
                fill: true
            }]
        },
        options: {
            responsive: true,
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
    
    // Users Chart
    const usersCtx = document.getElementById('users-chart').getContext('2d');
    new Chart(usersCtx, {
        type: 'bar',
        data: {
            labels: usersData.labels,
            datasets: [{
                label: 'New Users',
                data: usersData.data,
                backgroundColor: '#764ba2'
            }]
        },
        options: {
            responsive: true,
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
}

async function loadSystemInfo() {
    try {
        const response = await fetch('/api/system-health', {
            headers: setAuthHeaders()
        });
        
        if (!response.ok) {
            throw new Error('Failed to fetch system info');
        }
        
        const systemInfo = await response.json();
        updateSystemInfo(systemInfo);
    } catch (error) {
        console.error('Error loading system info:', error);
        showNotification('Error loading system info', 'error');
    }
}

function updateSystemInfo(info) {
    document.getElementById('system-status').textContent = info.status || 'Unknown';
    document.getElementById('db-status').textContent = info.database_status || 'Unknown';
    document.getElementById('uptime').textContent = info.uptime || 'Unknown';
    document.getElementById('memory-usage').textContent = info.memory_usage || 'Unknown';
}

// Modal functions
function showModal(modalId) {
    document.getElementById(modalId).style.display = 'block';
}

function closeModal(modalId) {
    document.getElementById(modalId).style.display = 'none';
}

function showUserActions(userId) {
    currentUserId = userId;
    showModal('user-actions-modal');
}

function showCreatePlanModal() {
    showModal('create-plan-modal');
}

// API Functions
async function createPlan() {
    const planData = {
        name: document.getElementById('plan-name').value,
        price_cents: parseInt(document.getElementById('plan-price').value),
        duration_days: parseInt(document.getElementById('plan-duration').value),
        max_groups: parseInt(document.getElementById('plan-max-groups').value),
        currency: document.getElementById('plan-currency').value
    };
    
    try {
        const response = await fetch('/api/plans', {
            method: 'POST',
            headers: setAuthHeaders(),
            body: JSON.stringify(planData)
        });
        
        if (!response.ok) {
            throw new Error('Failed to create plan');
        }
        
        closeModal('create-plan-modal');
        loadPlans();
        showNotification('Plan created successfully', 'success');
    } catch (error) {
        console.error('Error creating plan:', error);
        showNotification('Error creating plan', 'error');
    }
}

async function grantSubscription() {
    if (!currentUserId) return;
    
    try {
        const response = await fetch(`/api/users/${currentUserId}/grant`, {
            method: 'POST',
            headers: setAuthHeaders(),
            body: JSON.stringify({ plan_id: 2 }) // Premium plan
        });
        
        if (!response.ok) {
            throw new Error('Failed to grant subscription');
        }
        
        closeModal('user-actions-modal');
        loadUsers();
        showNotification('Subscription granted successfully', 'success');
    } catch (error) {
        console.error('Error granting subscription:', error);
        showNotification('Error granting subscription', 'error');
    }
}

async function revokeSubscription() {
    if (!currentUserId) return;
    
    try {
        const response = await fetch(`/api/users/${currentUserId}/revoke`, {
            method: 'POST',
            headers: setAuthHeaders()
        });
        
        if (!response.ok) {
            throw new Error('Failed to revoke subscription');
        }
        
        closeModal('user-actions-modal');
        loadUsers();
        showNotification('Subscription revoked successfully', 'success');
    } catch (error) {
        console.error('Error revoking subscription:', error);
        showNotification('Error revoking subscription', 'error');
    }
}

// Notification system
function showNotification(message, type) {
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.remove();
    }, 3000);
}

// Close modals when clicking outside
window.onclick = function(event) {
    const modals = document.querySelectorAll('.modal');
    modals.forEach(modal => {
        if (event.target === modal) {
            modal.style.display = 'none';
        }
    });
}

// Logout function
function logout() {
    localStorage.removeItem('authToken');
    window.location.href = '/';
}