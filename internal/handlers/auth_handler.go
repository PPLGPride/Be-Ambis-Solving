package handlers

import (
	"context"
	"time"

	"github.com/PPLGPride/Be-Ambis-Solving/internal/config"
	"github.com/PPLGPride/Be-Ambis-Solving/internal/services"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	Auth  services.AuthService
	Users services.UserService
}

func NewAuthHandler(a services.AuthService, u services.UserService) *AuthHandler {
	return &AuthHandler{Auth: a, Users: u}
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	token, uid, err := h.Auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid email or password"})
	}
	return c.JSON(fiber.Map{"token": token, "userId": uid})
}

type registerReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// (Dev convenience) enable only if ENABLE_REGISTER=true
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	if !config.Cfg.EnableRegister {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "registration disabled"})
	}
	var req registerReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	u, err := h.Users.Create(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":    u.ID.Hex(),
		"email": u.Email,
		"name":  u.Name,
	})
}
