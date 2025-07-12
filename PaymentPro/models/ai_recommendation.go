package models

import (
	"time"
)

// AIRecommendation represents an AI-powered recommendation for bot behavior
type AIRecommendation struct {
	ID               int       `json:"id" db:"id"`
	GroupID          *int64    `json:"group_id" db:"group_id"`
	RecommendationType string  `json:"recommendation_type" db:"recommendation_type"`
	Title            string    `json:"title" db:"title"`
	Description      string    `json:"description" db:"description"`
	Severity         string    `json:"severity" db:"severity"` // low, medium, high, critical
	Status           string    `json:"status" db:"status"`     // pending, implemented, dismissed
	Confidence       float64   `json:"confidence" db:"confidence"`
	AnalysisData     string    `json:"analysis_data" db:"analysis_data"` // JSON with supporting data
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	ExpiresAt        *time.Time `json:"expires_at" db:"expires_at"`
}

// BehaviorAnalytics represents analytics data for AI recommendations
type BehaviorAnalytics struct {
	ID              int       `json:"id" db:"id"`
	GroupID         int64     `json:"group_id" db:"group_id"`
	AnalysisType    string    `json:"analysis_type" db:"analysis_type"` // activity, moderation, engagement
	MetricName      string    `json:"metric_name" db:"metric_name"`
	MetricValue     float64   `json:"metric_value" db:"metric_value"`
	PreviousValue   *float64  `json:"previous_value" db:"previous_value"`
	ChangePercent   *float64  `json:"change_percent" db:"change_percent"`
	TimeWindow      string    `json:"time_window" db:"time_window"` // 1h, 24h, 7d, 30d
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// RecommendationEngine represents the AI engine configuration
type RecommendationEngine struct {
	ID                  int       `json:"id" db:"id"`
	Name                string    `json:"name" db:"name"`
	Version             string    `json:"version" db:"version"`
	IsActive            bool      `json:"is_active" db:"is_active"`
	ModelConfiguration  string    `json:"model_configuration" db:"model_configuration"` // JSON config
	LastTrainingDate    *time.Time `json:"last_training_date" db:"last_training_date"`
	AccuracyScore       *float64  `json:"accuracy_score" db:"accuracy_score"`
	RecommendationsCount int      `json:"recommendations_count" db:"recommendations_count"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// GroupBehaviorPattern represents detected patterns in group behavior
type GroupBehaviorPattern struct {
	ID            int       `json:"id" db:"id"`
	GroupID       int64     `json:"group_id" db:"group_id"`
	PatternType   string    `json:"pattern_type" db:"pattern_type"` // spam_surge, quiet_period, high_activity
	PatternData   string    `json:"pattern_data" db:"pattern_data"` // JSON with pattern details
	Confidence    float64   `json:"confidence" db:"confidence"`
	StartTime     time.Time `json:"start_time" db:"start_time"`
	EndTime       *time.Time `json:"end_time" db:"end_time"`
	IsOngoing     bool      `json:"is_ongoing" db:"is_ongoing"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}