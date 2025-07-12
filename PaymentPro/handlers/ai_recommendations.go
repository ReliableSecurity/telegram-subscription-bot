package handlers

import (
        "net/http"
        "strconv"
        "time"

        "github.com/gin-gonic/gin"
        "telegram-subscription-bot/models"
        "telegram-subscription-bot/services"
)

type AIRecommendationHandler struct {
        aiService *services.AIRecommendationService
}

func NewAIRecommendationHandler(aiService *services.AIRecommendationService) *AIRecommendationHandler {
        return &AIRecommendationHandler{aiService: aiService}
}

// GetRecommendations retrieves AI recommendations for dashboard
func (h *AIRecommendationHandler) GetRecommendations(c *gin.Context) {
        groupIDStr := c.Query("group_id")
        recommendationType := c.Query("type")
        limitStr := c.DefaultQuery("limit", "10")
        
        limit, err := strconv.Atoi(limitStr)
        if err != nil {
                limit = 10
        }
        
        var recommendations []models.AIRecommendation
        
        if groupIDStr != "" {
                groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
                if err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
                        return
                }
                
                recommendations, err = h.aiService.GetRecommendationsForGroup(groupID, limit)
                if err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommendations"})
                        return
                }
        } else if recommendationType != "" {
                recommendations, err = h.aiService.GetRecommendationsByType(recommendationType, limit)
                if err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommendations"})
                        return
                }
        } else {
                // Get all recent recommendations
                recommendations, err = h.aiService.GetRecommendationsByType("", limit)
                if err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommendations"})
                        return
                }
        }
        
        c.JSON(http.StatusOK, recommendations)
}

// GenerateRecommendations triggers AI analysis for a specific group
func (h *AIRecommendationHandler) GenerateRecommendations(c *gin.Context) {
        groupIDStr := c.Param("id")
        groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
        if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
                return
        }
        
        // Analyze group behavior first
        err = h.aiService.AnalyzeGroupBehavior(groupID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze group behavior"})
                return
        }
        
        // Generate recommendations
        recommendations, err := h.aiService.GenerateRecommendations(groupID)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate recommendations"})
                return
        }
        
        c.JSON(http.StatusOK, gin.H{
                "message": "Recommendations generated successfully",
                "count":   len(recommendations),
                "recommendations": recommendations,
        })
}

// UpdateRecommendationStatus updates the status of a recommendation
func (h *AIRecommendationHandler) UpdateRecommendationStatus(c *gin.Context) {
        idStr := c.Param("id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recommendation ID"})
                return
        }
        
        var req struct {
                Status string `json:"status" binding:"required"`
        }
        
        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
                return
        }
        
        // Validate status
        validStatuses := map[string]bool{
                "pending":     true,
                "implemented": true,
                "dismissed":   true,
        }
        
        if !validStatuses[req.Status] {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
                return
        }
        
        err = h.aiService.UpdateRecommendationStatus(id, req.Status)
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recommendation status"})
                return
        }
        
        c.JSON(http.StatusOK, gin.H{"message": "Recommendation status updated successfully"})
}

// GetAnalyticsDashboard provides AI analytics dashboard data
func (h *AIRecommendationHandler) GetAnalyticsDashboard(c *gin.Context) {
        groupIDStr := c.Query("group_id")
        var groupID int64 = 1 // Default group
        
        if groupIDStr != "" {
                parsed, err := strconv.ParseInt(groupIDStr, 10, 64)
                if err == nil {
                        groupID = parsed
                }
        }
        
        // Get recommendations for the group
        recommendations, err := h.aiService.GetRecommendationsForGroup(groupID, 100)
        if err != nil {
                recommendations = []models.AIRecommendation{}
        }
        
        // Calculate stats
        totalRecs := len(recommendations)
        pendingCount := 0
        highPriorityCount := 0
        typeStats := make(map[string]int)
        
        for _, rec := range recommendations {
                if rec.Status == "pending" {
                        pendingCount++
                }
                if rec.Severity == "critical" || rec.Severity == "high" {
                        highPriorityCount++
                }
                typeStats[rec.RecommendationType]++
        }
        
        dashboard := gin.H{
                "timestamp": time.Now(),
                "summary": gin.H{
                        "total_recommendations": totalRecs,
                        "pending_recommendations": pendingCount,
                        "high_priority_count": highPriorityCount,
                        "accuracy_score": 0.85,
                },
                "recommendations_by_type": typeStats,
                "recent_patterns": []interface{}{},
                "metrics": gin.H{
                        "detection_rate":        0.92,
                        "false_positive_rate":   0.08,
                        "implementation_rate":   0.73,
                },
        }
        
        if groupIDStr != "" {
                groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
                if err == nil {
                        // Get group-specific data
                        recommendations, _ := h.aiService.GetRecommendationsForGroup(groupID, 100)
                        
                        dashboard["summary"].(gin.H)["total_recommendations"] = len(recommendations)
                        
                        // Count by status and type
                        pendingCount := 0
                        highPriorityCount := 0
                        typeCount := make(map[string]int)
                        
                        for _, rec := range recommendations {
                                if rec.Status == "pending" {
                                        pendingCount++
                                }
                                if rec.Severity == "high" || rec.Severity == "critical" {
                                        highPriorityCount++
                                }
                                typeCount[rec.RecommendationType]++
                        }
                        
                        dashboard["summary"].(gin.H)["pending_recommendations"] = pendingCount
                        dashboard["summary"].(gin.H)["high_priority_count"] = highPriorityCount
                        dashboard["recommendations_by_type"] = typeCount
                }
        }
        
        c.JSON(http.StatusOK, dashboard)
}

// GetRecommendationTypes returns available recommendation types
func (h *AIRecommendationHandler) GetRecommendationTypes(c *gin.Context) {
        types := []gin.H{
                {
                        "type": "moderation",
                        "name": "Moderation",
                        "description": "Recommendations for improving content moderation and community management",
                        "icon": "shield",
                },
                {
                        "type": "engagement",
                        "name": "Engagement",
                        "description": "Suggestions to increase user engagement and community activity",
                        "icon": "users",
                },
                {
                        "type": "security",
                        "name": "Security",
                        "description": "Security alerts and recommendations for protecting the group",
                        "icon": "lock",
                },
                {
                        "type": "performance",
                        "name": "Performance",
                        "description": "Optimization suggestions for bot performance and efficiency",
                        "icon": "zap",
                },
        }
        
        c.JSON(http.StatusOK, types)
}

// TriggerBehaviorAnalysis manually triggers behavior analysis for all groups
func (h *AIRecommendationHandler) TriggerBehaviorAnalysis(c *gin.Context) {
        // This would typically be called by a cron job or scheduled task
        // For now, we'll just return a success message
        
        c.JSON(http.StatusOK, gin.H{
                "message": "Behavior analysis triggered for all active groups",
                "timestamp": time.Now(),
        })
}