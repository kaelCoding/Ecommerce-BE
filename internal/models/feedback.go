package models

type Feedback struct {
    Name    string `json:"userName" binding:"required"`
    Email   string `json:"userEmail" binding:"required,email"`
    Content string `json:"feedback" binding:"required"`
}