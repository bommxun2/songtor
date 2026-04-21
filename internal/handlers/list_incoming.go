package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ListIncomingHandler struct {
	db *gorm.DB
}

func NewListIncomingHandler(db *gorm.DB) *ListIncomingHandler {
	return &ListIncomingHandler{db: db}
}

func (h *ListIncomingHandler) ListIncomingPatients(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}
