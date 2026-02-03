package model

import "github.com/gofiber/fiber/v2"

// ─── Standard API Response ──────────────────────────────────────────────────

type APIResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
}

// ─── Pagination Meta ────────────────────────────────────────────────────────

type PaginationMeta struct {
	TotalData  int `json:"total_data"`
	TotalPage  int `json:"total_page"`
	CurrentPage int `json:"current_page"`
	PerPage    int `json:"per_page"`
}

type PaginatedResponse struct {
	Status     bool           `json:"status"`
	Message    string         `json:"message"`
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// ─── Helper Functions ────────────────────────────────────────────────────────

func SuccessResponse(c *fiber.Ctx, code int, message string, data interface{}) error {
	return c.Status(code).JSON(APIResponse{
		Status:  true,
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *fiber.Ctx, code int, message string, errors ...[]string) error {
	resp := APIResponse{
		Status:  false,
		Message: message,
	}
	if len(errors) > 0 && len(errors[0]) > 0 {
		resp.Errors = errors[0]
	}
	return c.Status(code).JSON(resp)
}

func PaginatedSuccessResponse(c *fiber.Ctx, data interface{}, totalData, currentPage, perPage int) error {
	totalPage := totalData / perPage
	if totalData%perPage != 0 {
		totalPage++
	}
	if totalPage == 0 {
		totalPage = 1
	}

	return c.Status(200).JSON(PaginatedResponse{
		Status:  true,
		Message: "Berhasil",
		Data:    data,
		Pagination: PaginationMeta{
			TotalData:   totalData,
			TotalPage:   totalPage,
			CurrentPage: currentPage,
			PerPage:     perPage,
		},
	})
}
