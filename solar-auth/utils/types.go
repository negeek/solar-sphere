package utils
import (
	"time"
	"github.com/golang-jwt/jwt/v5"
)

type Response struct {
	StatusCode	int			`json:"statuscode"`
	Success  bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}


type UserClaim struct {
	jwt.RegisteredClaims
	Email string
	DateTime  time.Time
}