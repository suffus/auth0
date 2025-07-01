package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/jackc/pgtype"
)

type UserActivityService struct {
	db *gorm.DB
}

func NewUserActivityService(db *gorm.DB) *UserActivityService {
	return &UserActivityService{db: db}
}

// ActivityFilter represents the filters for querying user activity
type ActivityFilter struct {
	FromDateTime *time.Time
	ToDateTime   *time.Time
	UserIDs      []uuid.UUID
	LocationIDs  []uuid.UUID
	StatusIDs    []uuid.UUID
	ActionIDs    []uuid.UUID
	Limit        int
	Offset       int
}

// ActivitySummary represents a summary of user activity
type ActivitySummary struct {
	UserID       uuid.UUID `json:"user_id"`
	UserName     string    `json:"user_name"`
	TotalHours   float64   `json:"total_hours"`
	BreakHours   float64   `json:"break_hours"`
	WorkHours    float64   `json:"work_hours"`
	MeetingHours float64   `json:"meeting_hours"`
	SignIns      int       `json:"sign_ins"`
	SignOuts     int       `json:"sign_outs"`
}

// GetUserActivity retrieves user activity history with filters
func (s *UserActivityService) GetUserActivity(filter ActivityFilter) ([]database.UserActivityHistory, int64, error) {
	var activities []database.UserActivityHistory
	var total int64

	query := s.db.Model(&database.UserActivityHistory{}).
		Preload("User").
		Preload("Action").
		Preload("Location").
		Preload("Status")

	// Apply filters
	query = s.applyFilters(query, filter)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count activities: %w", err)
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Order by from_datetime descending
	query = query.Order("from_datetime DESC")

	// Execute query
	if err := query.Find(&activities).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get activities: %w", err)
	}

	return activities, total, nil
}

// GetActivityByUser retrieves activity for a specific user
func (s *UserActivityService) GetActivityByUser(userID uuid.UUID, filter ActivityFilter) ([]database.UserActivityHistory, int64, error) {
	filter.UserIDs = []uuid.UUID{userID}
	return s.GetUserActivity(filter)
}

// GetActivityByLocation retrieves activity for specific locations
func (s *UserActivityService) GetActivityByLocation(locationIDs []uuid.UUID, filter ActivityFilter) ([]database.UserActivityHistory, int64, error) {
	filter.LocationIDs = locationIDs
	return s.GetUserActivity(filter)
}

// GetActivityByStatus retrieves activity for specific statuses
func (s *UserActivityService) GetActivityByStatus(statusIDs []uuid.UUID, filter ActivityFilter) ([]database.UserActivityHistory, int64, error) {
	filter.StatusIDs = statusIDs
	return s.GetUserActivity(filter)
}

// GetActivityByAction retrieves activity for specific actions
func (s *UserActivityService) GetActivityByAction(actionIDs []uuid.UUID, filter ActivityFilter) ([]database.UserActivityHistory, int64, error) {
	filter.ActionIDs = actionIDs
	return s.GetUserActivity(filter)
}

// GetActivitySummary retrieves activity summary for users
func (s *UserActivityService) GetActivitySummary(userIDs []uuid.UUID, fromTime, toTime time.Time) ([]ActivitySummary, error) {
	var summaries []ActivitySummary

	// Build the base query
	query := `
		SELECT 
			u.id as user_id,
			CONCAT(u.first_name, ' ', u.last_name) as user_name,
			COALESCE(SUM(
				CASE 
					WHEN a.name IN ('work-start', 'work-end', 'meeting-start', 'meeting-end') 
					THEN EXTRACT(EPOCH FROM (COALESCE(uah.to_datetime, NOW()) - uah.from_datetime)) / 3600
					ELSE 0 
				END
			), 0) as total_hours,
			COALESCE(SUM(
				CASE 
					WHEN a.name IN ('break-start', 'break-end') 
					THEN EXTRACT(EPOCH FROM (COALESCE(uah.to_datetime, NOW()) - uah.from_datetime)) / 3600
					ELSE 0 
				END
			), 0) as break_hours,
			COALESCE(SUM(
				CASE 
					WHEN a.name IN ('work-start', 'work-end') 
					THEN EXTRACT(EPOCH FROM (COALESCE(uah.to_datetime, NOW()) - uah.from_datetime)) / 3600
					ELSE 0 
				END
			), 0) as work_hours,
			COALESCE(SUM(
				CASE 
					WHEN a.name IN ('meeting-start', 'meeting-end') 
					THEN EXTRACT(EPOCH FROM (COALESCE(uah.to_datetime, NOW()) - uah.from_datetime)) / 3600
					ELSE 0 
				END
			), 0) as meeting_hours,
			COUNT(CASE WHEN a.name = 'user-signin' THEN 1 END) as sign_ins,
			COUNT(CASE WHEN a.name = 'user-signout' THEN 1 END) as sign_outs
		FROM users u
		LEFT JOIN user_activity_history uah ON u.id = uah.user_id
		LEFT JOIN actions a ON uah.action_id = a.id
		WHERE uah.from_datetime >= ? AND uah.from_datetime <= ?
	`

	var args []interface{}
	args = append(args, fromTime, toTime)

	if len(userIDs) > 0 {
		placeholders := make([]string, len(userIDs))
		for i := range userIDs {
			placeholders[i] = "?"
			args = append(args, userIDs[i])
		}
		query += fmt.Sprintf(" AND u.id IN (%s)", strings.Join(placeholders, ","))
	}

	query += `
		GROUP BY u.id, u.first_name, u.last_name
		ORDER BY u.first_name, u.last_name
	`

	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to execute summary query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var summary ActivitySummary
		err := rows.Scan(
			&summary.UserID,
			&summary.UserName,
			&summary.TotalHours,
			&summary.BreakHours,
			&summary.WorkHours,
			&summary.MeetingHours,
			&summary.SignIns,
			&summary.SignOuts,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan summary row: %w", err)
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// applyFilters applies the given filters to the query
func (s *UserActivityService) applyFilters(query *gorm.DB, filter ActivityFilter) *gorm.DB {
	if filter.FromDateTime != nil {
		query = query.Where("from_datetime >= ?", filter.FromDateTime)
	}

	if filter.ToDateTime != nil {
		query = query.Where("from_datetime <= ?", filter.ToDateTime)
	}

	if len(filter.UserIDs) > 0 {
		query = query.Where("user_id IN ?", filter.UserIDs)
	}

	if len(filter.LocationIDs) > 0 {
		query = query.Where("location_id IN ?", filter.LocationIDs)
	}

	if len(filter.StatusIDs) > 0 {
		query = query.Where("status_id IN ?", filter.StatusIDs)
	}

	if len(filter.ActionIDs) > 0 {
		query = query.Where("action_id IN ?", filter.ActionIDs)
	}

	return query
}

// CreateActivity creates a new activity record
func (s *UserActivityService) CreateActivity(activity *database.UserActivityHistory) error {
	return s.db.Create(activity).Error
}

// CreateUserActivity creates a new user activity history record
// user, status, and action are required (pointers to objects)
// location is optional (can be nil)
// details is optional JSON data
// closePreviousActivity if true, will close the user's most recent open activity
func (s *UserActivityService) CreateUserActivity(
	user *database.User,
	status *database.UserStatus,
	action *database.Action,
	location *database.Location,
	details map[string]interface{},
	closePreviousActivity bool,
) (*database.UserActivityHistory, error) {
	// Validate required fields
	if user == nil {
		return nil, fmt.Errorf("user is required")
	}
	if status == nil {
		return nil, fmt.Errorf("status is required")
	}
	if action == nil {
		return nil, fmt.Errorf("action is required")
	}

	// Set default details if nil
	if details == nil {
		details = make(map[string]interface{})
	}

	// Get current time for FromDateTime
	now := time.Now()

	// If closePreviousActivity is true, close the user's most recent open activity
	if closePreviousActivity {
		err := s.closeUserPreviousActivity(user.ID, now)
		if err != nil {
			return nil, fmt.Errorf("failed to close previous activity: %w", err)
		}
	}

	// Create the new activity record
	activity := &database.UserActivityHistory{
		ID:           uuid.New(),
		UserID:       user.ID,
		StatusID:     &status.ID,
		ActionID:     action.ID,
		FromDateTime: now,
		ToDateTime:   nil, // Will be set when this activity is closed
		Details:      pgtype.JSONB{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Set location ID if provided
	if location != nil {
		activity.LocationID = &location.ID
	}

	// Convert details map to JSONB
	if len(details) > 0 {
		if err := activity.Details.Set(details); err != nil {
			return nil, fmt.Errorf("failed to marshal details: %w", err)
		}
	} else {
		// Set empty JSON object if no details
		activity.Details = pgtype.JSONB{
			Bytes:  []byte("{}"),
			Status: pgtype.Present,
		}
	}

	// Save to database
	if err := s.db.Create(activity).Error; err != nil {
		return nil, fmt.Errorf("failed to create user activity: %w", err)
	}

	return activity, nil
}

// closeUserPreviousActivity closes the user's most recent open activity
// by setting its ToDateTime to the provided closeTime
func (s *UserActivityService) closeUserPreviousActivity(userID uuid.UUID, closeTime time.Time) error {
	// Find the most recent open activity for this user
	var previousActivity database.UserActivityHistory
	err := s.db.Where("user_id = ? AND to_datetime IS NULL", userID).
		Order("from_datetime DESC").
		First(&previousActivity).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No open activity found, which is fine
			return nil
		}
		return fmt.Errorf("failed to find previous activity: %w", err)
	}

	// Close the previous activity
	previousActivity.ToDateTime = &closeTime
	previousActivity.UpdatedAt = closeTime

	if err := s.db.Save(&previousActivity).Error; err != nil {
		return fmt.Errorf("failed to close previous activity: %w", err)
	}

	return nil
}

// CloseUserActivity closes a specific user activity by setting its ToDateTime
func (s *UserActivityService) CloseUserActivity(activityID uuid.UUID, closeTime time.Time) error {
	var activity database.UserActivityHistory
	err := s.db.Where("id = ?", activityID).First(&activity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("activity not found")
		}
		return fmt.Errorf("failed to find activity: %w", err)
	}

	// Check if activity is already closed
	if activity.ToDateTime != nil {
		return fmt.Errorf("activity is already closed")
	}

	// Close the activity
	activity.ToDateTime = &closeTime
	activity.UpdatedAt = closeTime

	if err := s.db.Save(&activity).Error; err != nil {
		return fmt.Errorf("failed to close activity: %w", err)
	}

	return nil
}

// UpdateActivity updates an existing activity record
func (s *UserActivityService) UpdateActivity(activity *database.UserActivityHistory) error {
	return s.db.Save(activity).Error
}

// GetActivityByID retrieves a specific activity by ID
func (s *UserActivityService) GetActivityByID(id uuid.UUID) (*database.UserActivityHistory, error) {
	var activity database.UserActivityHistory
	err := s.db.Preload("User").
		Preload("Action").
		Preload("Location").
		Preload("Status").
		Where("id = ?", id).
		First(&activity).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	return &activity, nil
} 