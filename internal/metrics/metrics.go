package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	LoginAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sso_login_attempts_total",
		Help: "Total number of login attempts",
	}, []string{"status"})

	RegisterAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sso_register_attempts_total",
		Help: "Total number of registration attempts",
	}, []string{"status"})

	TokenValidations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sso_token_validation_total",
		Help: "Total number of token validations",
	}, []string{"status"})

	TokenRotations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sso_token_rotation_total",
		Help: "Total number of refresh token rotations",
	}, []string{"status"})
)
