// Package metrics provides Prometheus metrics for the game.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ActiveRooms tracks the number of active game rooms.
	ActiveRooms = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "snake_active_rooms",
		Help: "number of active game rooms",
	})

	// ActivePlayers tracks the number of active players.
	ActivePlayers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "snake_active_players",
		Help: "number of active players across all rooms",
	})

	// RoomsCreated tracks total rooms created.
	RoomsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "snake_rooms_created_total",
		Help: "total number of rooms created",
	})

	// RoomsCleaned tracks total rooms cleaned up.
	RoomsCleaned = promauto.NewCounter(prometheus.CounterOpts{
		Name: "snake_rooms_cleaned_total",
		Help: "total number of rooms cleaned up",
	})

	// GameDuration tracks game duration in seconds.
	GameDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "snake_game_duration_seconds",
		Help:    "game duration in seconds",
		Buckets: prometheus.ExponentialBuckets(10, 2, 8), //nolint:mnd // 10s to ~20min
	})

	// RequestDuration tracks request latency by endpoint.
	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "snake_request_duration_seconds",
		Help:    "request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method"})

	// ErrorsTotal tracks errors by type.
	ErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "snake_errors_total",
		Help: "total number of errors",
	}, []string{"type"})

	// WebSocketConnections tracks active WebSocket connections.
	WebSocketConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "snake_websocket_connections",
		Help: "number of active websocket connections",
	})

	// MessagesReceived tracks messages received by type.
	MessagesReceived = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "snake_messages_received_total",
		Help: "total messages received",
	}, []string{"type"})

	// MessagesSent tracks messages sent by type.
	MessagesSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "snake_messages_sent_total",
		Help: "total messages sent",
	}, []string{"type"})
)
