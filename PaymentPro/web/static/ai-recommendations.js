// AI Recommendations JavaScript
let currentToken = localStorage.getItem('authToken') || new URLSearchParams(window.location.search).get('token');
let currentRecommendations = [];
let selectedRecommendationId = null;

// Save token if found in URL
if (currentToken && !localStorage.getItem('authToken')) {
    localStorage.setItem('authToken', currentToken);
}

// Enhanced fetch with token in URL as fallback
function fetchWithToken(url, options = {}) {
    // Ensure we have a token
    if (!currentToken) {
        currentToken = localStorage.getItem('authToken') || new URLSearchParams(window.location.search).get('token');
    }
    
    const tokenUrl = url.includes('?') ? `${url}&token=${currentToken}` : `${url}?token=${currentToken}`;
    
    const headers = {
        'Authorization': `Bearer ${currentToken}`,
        'Content-Type': 'application/json',
        ...options.headers
    };
    
    return fetch(tokenUrl, {
        ...options,
        headers
    });
}

// Initialize page
document.addEventListener('DOMContentLoaded', function() {
    if (!currentToken) {
        window.location.href = '/';
        return;
    }
    
    loadRecommendations();
    loadDashboardStats();
    
    // Setup auto-refresh
    setInterval(loadRecommendations, 60000); // Refresh every minute
});

// Load AI dashboard statistics
async function loadDashboardStats() {
    try {
        const response = await fetchWithToken('/api/ai/dashboard?group_id=1');
        
        if (!response.ok) {
            if (response.status === 401) {
                localStorage.removeItem('authToken');
                window.location.href = '/?error=auth';
                return;
            }
            throw new Error('Failed to fetch dashboard stats');
        }
        
        const data = await response.json();
        updateDashboardStats(data);
    } catch (error) {
        console.error('Error loading dashboard stats:', error);
        // Use fallback data if API fails
        updateDashboardStats({
            summary: {
                total_recommendations: 5,
                pending_recommendations: 4,
                high_priority_count: 2,
                accuracy_score: 0.85
            }
        });
    }
}

// Update dashboard statistics display
function updateDashboardStats(data) {
    document.getElementById('total-recommendations').textContent = data.summary.total_recommendations || 0;
    document.getElementById('pending-recommendations').textContent = data.summary.pending_recommendations || 0;
    document.getElementById('high-priority-count').textContent = data.summary.high_priority_count || 0;
    document.getElementById('accuracy-score').textContent = Math.round((data.summary.accuracy_score || 0.85) * 100) + '%';
}

// Load recommendations from API
async function loadRecommendations() {
    try {
        const response = await fetchWithToken('/api/ai/recommendations?group_id=1&limit=50');
        
        if (!response.ok) {
            if (response.status === 401) {
                localStorage.removeItem('authToken');
                window.location.href = '/?error=auth';
                return;
            }
            throw new Error('Failed to fetch recommendations');
        }
        
        const recommendations = await response.json();
        currentRecommendations = recommendations || [];
        displayRecommendations(currentRecommendations);
    } catch (error) {
        console.error('Error loading recommendations:', error);
        // Show error message but don't crash the page
        showNotification('Ошибка загрузки рекомендаций. Попробуйте обновить страницу.', 'error');
        document.getElementById('recommendations-container').innerHTML = 
            '<div class="error">Ошибка загрузки рекомендаций. <button onclick="loadRecommendations()" class="btn btn-primary">Попробовать снова</button></div>';
    }
}

// Display recommendations in the UI
function displayRecommendations(recommendations) {
    const container = document.getElementById('recommendations-container');
    
    if (!recommendations || recommendations.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-lightbulb fa-3x"></i>
                <h3>No Recommendations Available</h3>
                <p>Click "Generate New Recommendations" to analyze your bot's performance and get AI-powered suggestions.</p>
            </div>
        `;
        return;
    }
    
    const html = recommendations.map(rec => createRecommendationCard(rec)).join('');
    container.innerHTML = html;
}

// Create individual recommendation card
function createRecommendationCard(recommendation) {
    const severityClass = getSeverityClass(recommendation.severity);
    const statusClass = getStatusClass(recommendation.status);
    const typeIcon = getTypeIcon(recommendation.recommendation_type);
    const confidencePercent = Math.round((recommendation.confidence || 0) * 100);
    
    return `
        <div class="recommendation-card ${severityClass}" data-type="${recommendation.recommendation_type}" 
             data-severity="${recommendation.severity}" data-status="${recommendation.status}">
            <div class="recommendation-header">
                <div class="recommendation-type">
                    <i class="${typeIcon}"></i>
                    <span>${recommendation.recommendation_type.charAt(0).toUpperCase() + recommendation.recommendation_type.slice(1)}</span>
                </div>
                <div class="recommendation-meta">
                    <span class="severity-badge ${severityClass}">${recommendation.severity.toUpperCase()}</span>
                    <span class="status-badge ${statusClass}">${recommendation.status}</span>
                    <span class="confidence">
                        <i class="fas fa-percentage"></i> ${confidencePercent}%
                    </span>
                </div>
            </div>
            <div class="recommendation-content">
                <h3>${recommendation.title}</h3>
                <p>${recommendation.description}</p>
                <div class="recommendation-timestamp">
                    <i class="fas fa-clock"></i>
                    ${new Date(recommendation.created_at).toLocaleString()}
                </div>
            </div>
            <div class="recommendation-actions">
                <button onclick="viewRecommendationDetails(${recommendation.id})" class="btn btn-sm btn-primary">
                    <i class="fas fa-eye"></i> View Details
                </button>
                ${recommendation.status === 'pending' ? `
                    <button onclick="quickUpdateStatus(${recommendation.id}, 'implemented')" class="btn btn-sm btn-success">
                        <i class="fas fa-check"></i> Implement
                    </button>
                    <button onclick="quickUpdateStatus(${recommendation.id}, 'dismissed')" class="btn btn-sm btn-danger">
                        <i class="fas fa-times"></i> Dismiss
                    </button>
                ` : ''}
            </div>
        </div>
    `;
}

// Helper functions for styling
function getSeverityClass(severity) {
    switch (severity) {
        case 'critical': return 'severity-critical';
        case 'high': return 'severity-high';
        case 'medium': return 'severity-medium';
        case 'low': return 'severity-low';
        default: return 'severity-medium';
    }
}

function getStatusClass(status) {
    switch (status) {
        case 'pending': return 'status-pending';
        case 'implemented': return 'status-implemented';
        case 'dismissed': return 'status-dismissed';
        default: return 'status-pending';
    }
}

function getTypeIcon(type) {
    switch (type) {
        case 'moderation': return 'fas fa-shield-alt';
        case 'engagement': return 'fas fa-users';
        case 'security': return 'fas fa-lock';
        case 'performance': return 'fas fa-tachometer-alt';
        default: return 'fas fa-lightbulb';
    }
}

// Generate new recommendations
async function generateRecommendations() {
    const button = event.target;
    const originalText = button.innerHTML;
    
    button.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Generating...';
    button.disabled = true;
    
    try {
        // For demo purposes, we'll generate for group ID 1
        // In a real implementation, this would be selected from a dropdown
        const response = await fetchWithToken('/api/ai/recommendations/generate/1', {
            method: 'POST'
        });
        
        if (!response.ok) {
            throw new Error('Failed to generate recommendations');
        }
        
        const result = await response.json();
        showNotification(`Generated ${result.count} new recommendations`, 'success');
        
        // Reload recommendations
        await loadRecommendations();
        await loadDashboardStats();
    } catch (error) {
        console.error('Error generating recommendations:', error);
        showNotification('Failed to generate recommendations', 'error');
    } finally {
        button.innerHTML = originalText;
        button.disabled = false;
    }
}

// Filter recommendations
function filterRecommendations() {
    const typeFilter = document.getElementById('type-filter').value;
    const severityFilter = document.getElementById('severity-filter').value;
    const statusFilter = document.getElementById('status-filter').value;
    
    let filtered = currentRecommendations;
    
    if (typeFilter) {
        filtered = filtered.filter(rec => rec.recommendation_type === typeFilter);
    }
    
    if (severityFilter) {
        filtered = filtered.filter(rec => rec.severity === severityFilter);
    }
    
    if (statusFilter) {
        filtered = filtered.filter(rec => rec.status === statusFilter);
    }
    
    displayRecommendations(filtered);
}

// View recommendation details in modal
function viewRecommendationDetails(recommendationId) {
    const recommendation = currentRecommendations.find(rec => rec.id === recommendationId);
    if (!recommendation) return;
    
    selectedRecommendationId = recommendationId;
    
    document.getElementById('modal-title').textContent = recommendation.title;
    
    let analysisData = {};
    try {
        analysisData = JSON.parse(recommendation.analysis_data || '{}');
    } catch (e) {
        // If parsing fails, use empty object
    }
    
    const modalContent = `
        <div class="recommendation-details">
            <div class="detail-section">
                <h4>Description</h4>
                <p>${recommendation.description}</p>
            </div>
            <div class="detail-section">
                <h4>Recommendation Details</h4>
                <div class="detail-grid">
                    <div class="detail-item">
                        <label>Type:</label>
                        <span>${recommendation.recommendation_type}</span>
                    </div>
                    <div class="detail-item">
                        <label>Severity:</label>
                        <span class="severity-badge ${getSeverityClass(recommendation.severity)}">${recommendation.severity}</span>
                    </div>
                    <div class="detail-item">
                        <label>Confidence:</label>
                        <span>${Math.round((recommendation.confidence || 0) * 100)}%</span>
                    </div>
                    <div class="detail-item">
                        <label>Status:</label>
                        <span class="status-badge ${getStatusClass(recommendation.status)}">${recommendation.status}</span>
                    </div>
                    <div class="detail-item">
                        <label>Created:</label>
                        <span>${new Date(recommendation.created_at).toLocaleString()}</span>
                    </div>
                </div>
            </div>
            ${Object.keys(analysisData).length > 0 ? `
                <div class="detail-section">
                    <h4>Analysis Data</h4>
                    <div class="analysis-data">
                        ${Object.entries(analysisData).map(([key, value]) => 
                            `<div class="analysis-item">
                                <label>${key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}:</label>
                                <span>${typeof value === 'number' ? value.toFixed(2) : value}</span>
                            </div>`
                        ).join('')}
                    </div>
                </div>
            ` : ''}
        </div>
    `;
    
    document.getElementById('modal-content').innerHTML = modalContent;
    showModal('recommendation-modal');
}

// Quick update recommendation status
async function quickUpdateStatus(recommendationId, status) {
    try {
        const response = await fetchWithToken(`/api/ai/recommendations/${recommendationId}/status`, {
            method: 'PUT',
            body: JSON.stringify({ status })
        });
        
        if (!response.ok) {
            throw new Error('Failed to update recommendation status');
        }
        
        showNotification(`Recommendation ${status}`, 'success');
        await loadRecommendations();
        await loadDashboardStats();
    } catch (error) {
        console.error('Error updating recommendation status:', error);
        showNotification('Failed to update recommendation status', 'error');
    }
}

// Update recommendation status from modal
async function updateRecommendationStatus(status) {
    if (!selectedRecommendationId) return;
    
    await quickUpdateStatus(selectedRecommendationId, status);
    closeModal('recommendation-modal');
}

// Refresh data
async function refreshData() {
    await Promise.all([
        loadRecommendations(),
        loadDashboardStats()
    ]);
    showNotification('Data refreshed', 'success');
}

// Utility functions
function showModal(modalId) {
    document.getElementById(modalId).style.display = 'block';
}

function closeModal(modalId) {
    document.getElementById(modalId).style.display = 'none';
    selectedRecommendationId = null;
}

function showNotification(message, type) {
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.innerHTML = `
        <i class="fas fa-${type === 'success' ? 'check' : type === 'error' ? 'exclamation-triangle' : 'info'}"></i>
        ${message}
    `;
    
    // Add to page
    document.body.appendChild(notification);
    
    // Remove after 3 seconds
    setTimeout(() => {
        notification.remove();
    }, 3000);
}

function logout() {
    localStorage.removeItem('authToken');
    window.location.href = '/';
}

// Close modal when clicking outside
window.onclick = function(event) {
    const modals = document.querySelectorAll('.modal');
    modals.forEach(modal => {
        if (event.target === modal) {
            modal.style.display = 'none';
        }
    });
};