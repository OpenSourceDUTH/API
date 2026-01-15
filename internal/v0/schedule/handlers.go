package schedule

import (
	"API/internal/v0/common"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler initialization that holds the Repository database connection so we can save the data
type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) PostFood(c *gin.Context) {
	var f Food
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}
	if err := h.repo.CreateFood(f.Name); err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{err.Error()}))
		return
	}
	c.JSON(http.StatusCreated, common.CreateSuccessResponse(nil))
}

func (h *Handler) PostVersion(c *gin.Context) {
	var v ScheduleVersion
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}
	id, err := h.repo.CreateVersion(v.StartingDate, v.EndingDate, v.IsCurrent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{err.Error()}))
		return
	}
	c.JSON(http.StatusCreated, common.CreateSuccessResponse(gin.H{"id": id}))
}

func (h *Handler) PostSchedule(c *gin.Context) {
	var s ScheduleItem
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}
	if err := h.repo.CreateScheduleItem(s.VersionID, s.WeekNumber, s.DayNumber, s.MealType, s.DishIDs); err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{err.Error()}))
		return
	}
	c.JSON(http.StatusCreated, common.CreateSuccessResponse(nil))
}

func (h *Handler) PostAnnouncement(c *gin.Context) {
	var a Announcement
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
		return
	}
	id, err := h.repo.CreateAnnouncement(a.Type, a.Content, a.StartingDate, a.EndingDate, a.IsCurrent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse([]string{err.Error()}))
		return
	}
	c.JSON(http.StatusCreated, common.CreateSuccessResponse(gin.H{"id": id}))
}

func (h *Handler) GetSchedule(c *gin.Context) {
	allParameter := c.Query("all")
	dateParameter := c.Query("date")

	// Check
	if dateParameter != "" {
		parsedTime, err := time.Parse("02012006", dateParameter)
		if err != nil {
			c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{"Invalid date format. Please use DDMMYYYY"}))
			return
		}

		formatedDate := parsedTime.Format("2006-01-02")
		schedule, err := h.repo.GetDateSchedule(formatedDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, common.CreateErrorResponse([]string{err.Error()}))
			return
		}
		c.JSON(http.StatusOK, common.CreateSuccessResponse(schedule))
		return
	} else if allParameter == "true" {

	}
}
