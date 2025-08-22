package handlers

import (
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/config"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/utils"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type TimelineHandler struct{}

func NewTimelineHandler() *TimelineHandler { return &TimelineHandler{} }

func parseRange(c *fiber.Ctx) (time.Time, time.Time, error) {
	fromStr := c.Query("from", "")
	toStr := c.Query("to", "")
	var from, to time.Time
	var err error
	if fromStr == "" || toStr == "" {
		// default 14 hari ke depan
		now := time.Now().UTC()
		from = now.AddDate(0, 0, -1)
		to = now.AddDate(0, 0, 14)
		return from, to, nil
	}
	from, err = time.Parse(time.RFC3339, fromStr)
	if err != nil {
		from, err = time.Parse("2006-01-02", fromStr)
	}
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err = time.Parse(time.RFC3339, toStr)
	if err != nil {
		to, err = time.Parse("2006-01-02", toStr)
	}
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return from, to, nil
}

func (h *TimelineHandler) Get(c *fiber.Ctx) error {
	// optional boardId filter
	var boardFilter bson.M
	if bid := c.Query("boardId"); bid != "" {
		oid, err := utils.MustObjectID(bid)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid boardId"})
		}
		boardFilter = bson.M{"boardId": oid}
	} else {
		boardFilter = bson.M{}
	}

	from, to, err := parseRange(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid date range"})
	}

	// Tasks yang ada di rentang (pakai dueDate/startDate)
	taskFilter := bson.M{
		"$or": []bson.M{
			{"dueDate": bson.M{"$gte": from, "$lte": to}},
			{"startDate": bson.M{"$gte": from, "$lte": to}},
		},
	}
	for k, v := range boardFilter {
		taskFilter[k] = v
	}

	curT, err := config.MongoDB.Collection("tasks").Find(c.Context(), taskFilter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	var tasks []bson.M
	if err := curT.All(c.Context(), &tasks); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Notes pada timeline (onTimelineAt range)
	noteFilter := bson.M{"onTimelineAt": bson.M{"$gte": from, "$lte": to}}
	for k, v := range boardFilter {
		noteFilter[k] = v
	}
	curN, err := config.MongoDB.Collection("notes").Find(c.Context(), noteFilter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	var notes []bson.M
	if err := curN.All(c.Context(), &notes); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"range": fiber.Map{"from": from, "to": to},
		"tasks": tasks,
		"notes": notes,
	})
}
