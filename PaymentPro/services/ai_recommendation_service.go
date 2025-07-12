package services

import (
        "encoding/json"
        "fmt"
        "log"
        "math"
        "time"

        "telegram-subscription-bot/models"
        "database/sql"
)

type AIRecommendationService struct {
        db *sql.DB
}

func NewAIRecommendationService(db *sql.DB) *AIRecommendationService {
        return &AIRecommendationService{db: db}
}

// AnalysisResult represents the result of behavior analysis
type AnalysisResult struct {
        MetricName    string  `json:"metric_name"`
        CurrentValue  float64 `json:"current_value"`
        PreviousValue float64 `json:"previous_value"`
        ChangePercent float64 `json:"change_percent"`
        Trend         string  `json:"trend"` // increasing, decreasing, stable
}

// RecommendationContext contains context for generating recommendations
type RecommendationContext struct {
        GroupID       int64                      `json:"group_id"`
        TimeWindow    string                     `json:"time_window"`
        Analytics     []models.BehaviorAnalytics `json:"analytics"`
        Patterns      []models.GroupBehaviorPattern `json:"patterns"`
        ViolationRate float64                    `json:"violation_rate"`
        ActivityLevel string                     `json:"activity_level"`
}

// GenerateRecommendations analyzes group behavior and generates AI-powered recommendations
func (s *AIRecommendationService) GenerateRecommendations(groupID int64) ([]models.AIRecommendation, error) {
        // Gather behavioral data
        context, err := s.gatherAnalysisContext(groupID)
        if err != nil {
                return nil, fmt.Errorf("failed to gather analysis context: %v", err)
        }

        var recommendations []models.AIRecommendation

        // Generate different types of recommendations
        moderationRecs := s.generateModerationRecommendations(context)
        engagementRecs := s.generateEngagementRecommendations(context)
        securityRecs := s.generateSecurityRecommendations(context)
        performanceRecs := s.generatePerformanceRecommendations(context)

        recommendations = append(recommendations, moderationRecs...)
        recommendations = append(recommendations, engagementRecs...)
        recommendations = append(recommendations, securityRecs...)
        recommendations = append(recommendations, performanceRecs...)

        // Save recommendations to database
        for i := range recommendations {
                err := s.saveRecommendation(&recommendations[i])
                if err != nil {
                        log.Printf("Failed to save recommendation: %v", err)
                }
        }

        return recommendations, nil
}

// gatherAnalysisContext collects all necessary data for analysis
func (s *AIRecommendationService) gatherAnalysisContext(groupID int64) (*RecommendationContext, error) {
        context := &RecommendationContext{
                GroupID:    groupID,
                TimeWindow: "24h",
        }

        // Get recent analytics
        analytics, err := s.getRecentAnalytics(groupID, "24h")
        if err != nil {
                log.Printf("Failed to get analytics: %v", err)
        }
        context.Analytics = analytics

        // Get behavior patterns
        patterns, err := s.getActivePatterns(groupID)
        if err != nil {
                log.Printf("Failed to get patterns: %v", err)
        }
        context.Patterns = patterns

        // Calculate violation rate
        context.ViolationRate = s.calculateViolationRate(groupID)

        // Determine activity level
        context.ActivityLevel = s.determineActivityLevel(analytics)

        return context, nil
}

// generateModerationRecommendations creates recommendations for moderation optimization
func (s *AIRecommendationService) generateModerationRecommendations(context *RecommendationContext) []models.AIRecommendation {
        var recommendations []models.AIRecommendation

        // High violation rate recommendation
        if context.ViolationRate > 0.15 { // More than 15% violation rate
                analysisData, _ := json.Marshal(map[string]interface{}{
                        "violation_rate": context.ViolationRate,
                        "threshold":      0.15,
                        "suggestion":     "tighten_moderation_rules",
                })

                recommendations = append(recommendations, models.AIRecommendation{
                        GroupID:            &context.GroupID,
                        RecommendationType: "moderation",
                        Title:              "High Violation Rate Detected",
                        Description:        fmt.Sprintf("Violation rate is %.1f%%, consider tightening moderation rules or adding more moderators", context.ViolationRate*100),
                        Severity:           "high",
                        Status:             "pending",
                        Confidence:         0.85,
                        AnalysisData:       string(analysisData),
                        CreatedAt:          time.Now(),
                        UpdatedAt:          time.Now(),
                })
        }

        // Spam pattern detection
        for _, pattern := range context.Patterns {
                if pattern.PatternType == "spam_surge" && pattern.Confidence > 0.7 {
                        analysisData, _ := json.Marshal(map[string]interface{}{
                                "pattern_type": pattern.PatternType,
                                "confidence":   pattern.Confidence,
                                "suggestion":   "enable_anti_spam_mode",
                        })

                        recommendations = append(recommendations, models.AIRecommendation{
                                GroupID:            &context.GroupID,
                                RecommendationType: "security",
                                Title:              "Spam Surge Detected",
                                Description:        "AI detected unusual spam activity pattern. Consider enabling enhanced anti-spam measures temporarily.",
                                Severity:           "medium",
                                Status:             "pending",
                                Confidence:         pattern.Confidence,
                                AnalysisData:       string(analysisData),
                                CreatedAt:          time.Now(),
                                UpdatedAt:          time.Now(),
                        })
                }
        }

        return recommendations
}

// generateEngagementRecommendations creates recommendations for improving engagement
func (s *AIRecommendationService) generateEngagementRecommendations(context *RecommendationContext) []models.AIRecommendation {
        var recommendations []models.AIRecommendation

        if context.ActivityLevel == "low" {
                analysisData, _ := json.Marshal(map[string]interface{}{
                        "activity_level": context.ActivityLevel,
                        "suggestion":     "increase_engagement_features",
                })

                recommendations = append(recommendations, models.AIRecommendation{
                        GroupID:            &context.GroupID,
                        RecommendationType: "engagement",
                        Title:              "Low Group Activity",
                        Description:        "Group activity is below average. Consider enabling welcome messages, polls, or interactive features to boost engagement.",
                        Severity:           "low",
                        Status:             "pending",
                        Confidence:         0.7,
                        AnalysisData:       string(analysisData),
                        CreatedAt:          time.Now(),
                        UpdatedAt:          time.Now(),
                })
        }

        return recommendations
}

// generateSecurityRecommendations creates security-related recommendations
func (s *AIRecommendationService) generateSecurityRecommendations(context *RecommendationContext) []models.AIRecommendation {
        var recommendations []models.AIRecommendation

        // Check for suspicious patterns
        for _, pattern := range context.Patterns {
                if pattern.PatternType == "suspicious_activity" && pattern.Confidence > 0.8 {
                        analysisData, _ := json.Marshal(map[string]interface{}{
                                "pattern_type": pattern.PatternType,
                                "confidence":   pattern.Confidence,
                                "suggestion":   "review_security_settings",
                        })

                        recommendations = append(recommendations, models.AIRecommendation{
                                GroupID:            &context.GroupID,
                                RecommendationType: "security",
                                Title:              "Suspicious Activity Pattern",
                                Description:        "AI detected unusual activity pattern that may indicate security risks. Review recent member additions and message patterns.",
                                Severity:           "critical",
                                Status:             "pending",
                                Confidence:         pattern.Confidence,
                                AnalysisData:       string(analysisData),
                                CreatedAt:          time.Now(),
                                UpdatedAt:          time.Now(),
                        })
                }
        }

        return recommendations
}

// generatePerformanceRecommendations creates performance optimization recommendations
func (s *AIRecommendationService) generatePerformanceRecommendations(context *RecommendationContext) []models.AIRecommendation {
        var recommendations []models.AIRecommendation

        // Check for high message volume
        for _, analytics := range context.Analytics {
                if analytics.MetricName == "messages_per_hour" && analytics.MetricValue > 100 {
                        analysisData, _ := json.Marshal(map[string]interface{}{
                                "messages_per_hour": analytics.MetricValue,
                                "threshold":         100,
                                "suggestion":        "optimize_processing",
                        })

                        recommendations = append(recommendations, models.AIRecommendation{
                                GroupID:            &context.GroupID,
                                RecommendationType: "performance",
                                Title:              "High Message Volume",
                                Description:        fmt.Sprintf("Group is processing %.0f messages/hour. Consider optimizing bot response time and enabling rate limiting.", analytics.MetricValue),
                                Severity:           "medium",
                                Status:             "pending",
                                Confidence:         0.9,
                                AnalysisData:       string(analysisData),
                                CreatedAt:          time.Now(),
                                UpdatedAt:          time.Now(),
                        })
                }
        }

        return recommendations
}

// Helper functions
func (s *AIRecommendationService) getRecentAnalytics(groupID int64, timeWindow string) ([]models.BehaviorAnalytics, error) {
        var analytics []models.BehaviorAnalytics
        
        // For now, return empty analytics as these tables might not exist yet
        // In a production system, this would query the actual analytics tables
        
        return analytics, nil
}

func (s *AIRecommendationService) getActivePatterns(groupID int64) ([]models.GroupBehaviorPattern, error) {
        var patterns []models.GroupBehaviorPattern
        
        // For now, return empty patterns as these tables might not exist yet
        // In a production system, this would query the actual pattern tables
        
        return patterns, nil
}

func (s *AIRecommendationService) calculateViolationRate(groupID int64) float64 {
        var totalMessages, violations int
        
        // Simulate data for demo purposes - in production these would be real queries
        totalMessages = 100 // Sample data
        violations = 5      // Sample data
        
        if totalMessages == 0 {
                return 0
        }
        
        return float64(violations) / float64(totalMessages)
}

func (s *AIRecommendationService) determineActivityLevel(analytics []models.BehaviorAnalytics) string {
        var messageRate float64 = 0
        
        for _, a := range analytics {
                if a.MetricName == "messages_per_hour" {
                        messageRate = a.MetricValue
                        break
                }
        }
        
        if messageRate < 5 {
                return "low"
        } else if messageRate < 20 {
                return "medium"
        }
        return "high"
}

func (s *AIRecommendationService) saveRecommendation(rec *models.AIRecommendation) error {
        query := `INSERT INTO ai_recommendations (group_id, recommendation_type, title, description, severity, status, confidence, analysis_data, created_at, updated_at) 
                          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
        
        return s.db.QueryRow(query, rec.GroupID, rec.RecommendationType, rec.Title, rec.Description, 
                rec.Severity, rec.Status, rec.Confidence, rec.AnalysisData, rec.CreatedAt, rec.UpdatedAt).Scan(&rec.ID)
}

// UpdateRecommendationStatus updates the status of a recommendation
func (s *AIRecommendationService) UpdateRecommendationStatus(id int, status string) error {
        query := `UPDATE ai_recommendations SET status = $1, updated_at = NOW() WHERE id = $2`
        _, err := s.db.Exec(query, status, id)
        return err
}

// GetRecommendationsForGroup retrieves recommendations for a specific group
func (s *AIRecommendationService) GetRecommendationsForGroup(groupID int64, limit int) ([]models.AIRecommendation, error) {
        var recommendations []models.AIRecommendation
        
        query := `SELECT id, group_id, recommendation_type, title, description, severity, status, confidence, analysis_data, created_at, updated_at, expires_at FROM ai_recommendations WHERE group_id = $1 ORDER BY created_at DESC LIMIT $2`
        
        rows, err := s.db.Query(query, groupID, limit)
        if err != nil {
                return recommendations, err
        }
        defer rows.Close()
        
        for rows.Next() {
                var rec models.AIRecommendation
                err := rows.Scan(&rec.ID, &rec.GroupID, &rec.RecommendationType, &rec.Title, &rec.Description, &rec.Severity, &rec.Status, &rec.Confidence, &rec.AnalysisData, &rec.CreatedAt, &rec.UpdatedAt, &rec.ExpiresAt)
                if err != nil {
                        continue
                }
                recommendations = append(recommendations, rec)
        }
        
        return recommendations, nil
}

// GetRecommendationsByType retrieves recommendations by type
func (s *AIRecommendationService) GetRecommendationsByType(recommendationType string, limit int) ([]models.AIRecommendation, error) {
        var recommendations []models.AIRecommendation
        
        var query string
        var args []interface{}
        
        if recommendationType == "" {
                query = `SELECT id, group_id, recommendation_type, title, description, severity, status, confidence, analysis_data, created_at, updated_at, expires_at FROM ai_recommendations WHERE status = 'pending' ORDER BY severity DESC, confidence DESC LIMIT $1`
                args = []interface{}{limit}
        } else {
                query = `SELECT id, group_id, recommendation_type, title, description, severity, status, confidence, analysis_data, created_at, updated_at, expires_at FROM ai_recommendations WHERE recommendation_type = $1 AND status = 'pending' ORDER BY severity DESC, confidence DESC LIMIT $2`
                args = []interface{}{recommendationType, limit}
        }
        
        rows, err := s.db.Query(query, args...)
        if err != nil {
                return recommendations, err
        }
        defer rows.Close()
        
        for rows.Next() {
                var rec models.AIRecommendation
                err := rows.Scan(&rec.ID, &rec.GroupID, &rec.RecommendationType, &rec.Title, &rec.Description, &rec.Severity, &rec.Status, &rec.Confidence, &rec.AnalysisData, &rec.CreatedAt, &rec.UpdatedAt, &rec.ExpiresAt)
                if err != nil {
                        continue
                }
                recommendations = append(recommendations, rec)
        }
        
        return recommendations, nil
}

// AnalyzeGroupBehavior performs real-time analysis and updates behavior analytics
func (s *AIRecommendationService) AnalyzeGroupBehavior(groupID int64) error {
        // Calculate various metrics
        metrics := s.calculateBehaviorMetrics(groupID)
        
        // Save analytics to database
        for metricName, value := range metrics {
                analytics := models.BehaviorAnalytics{
                        GroupID:      groupID,
                        AnalysisType: "behavior",
                        MetricName:   metricName,
                        MetricValue:  value,
                        TimeWindow:   "1h",
                        CreatedAt:    time.Now(),
                }
                
                s.saveBehaviorAnalytics(&analytics)
        }
        
        // Detect patterns
        s.detectBehaviorPatterns(groupID, metrics)
        
        return nil
}

func (s *AIRecommendationService) calculateBehaviorMetrics(groupID int64) map[string]float64 {
        metrics := make(map[string]float64)
        
        // For demo purposes, simulate realistic metrics
        // In production, these would be real database queries
        metrics["messages_per_hour"] = 15.0 + float64(groupID%10) // Vary by group
        metrics["unique_users_per_hour"] = 8.0 + float64(groupID%5)
        metrics["avg_message_length"] = 45.0 + float64(groupID%20)
        metrics["violations_per_hour"] = 2.0 + float64(groupID%3)
        
        return metrics
}

func (s *AIRecommendationService) detectBehaviorPatterns(groupID int64, metrics map[string]float64) {
        // Detect spam surge
        if metrics["messages_per_hour"] > 50 && metrics["avg_message_length"] < 20 {
                confidence := math.Min(0.9, metrics["messages_per_hour"]/100)
                s.savePattern(groupID, "spam_surge", confidence, map[string]interface{}{
                        "messages_per_hour": metrics["messages_per_hour"],
                        "avg_length":        metrics["avg_message_length"],
                })
        }
        
        // Detect quiet period
        if metrics["messages_per_hour"] < 2 && metrics["unique_users_per_hour"] < 2 {
                s.savePattern(groupID, "quiet_period", 0.8, map[string]interface{}{
                        "messages_per_hour": metrics["messages_per_hour"],
                        "unique_users":      metrics["unique_users_per_hour"],
                })
        }
}

func (s *AIRecommendationService) saveBehaviorAnalytics(analytics *models.BehaviorAnalytics) error {
        query := `INSERT INTO behavior_analytics (group_id, analysis_type, metric_name, metric_value, time_window, created_at) 
                          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
        
        return s.db.QueryRow(query, analytics.GroupID, analytics.AnalysisType, analytics.MetricName, 
                analytics.MetricValue, analytics.TimeWindow, analytics.CreatedAt).Scan(&analytics.ID)
}

func (s *AIRecommendationService) savePattern(groupID int64, patternType string, confidence float64, data map[string]interface{}) error {
        patternData, _ := json.Marshal(data)
        
        pattern := models.GroupBehaviorPattern{
                GroupID:     groupID,
                PatternType: patternType,
                PatternData: string(patternData),
                Confidence:  confidence,
                StartTime:   time.Now(),
                IsOngoing:   true,
                CreatedAt:   time.Now(),
        }
        
        query := `INSERT INTO group_behavior_patterns (group_id, pattern_type, pattern_data, confidence, start_time, is_ongoing, created_at) 
                          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
        
        return s.db.QueryRow(query, pattern.GroupID, pattern.PatternType, pattern.PatternData, 
                pattern.Confidence, pattern.StartTime, pattern.IsOngoing, pattern.CreatedAt).Scan(&pattern.ID)
}