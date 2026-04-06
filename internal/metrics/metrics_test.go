package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestActiveRooms(t *testing.T) {
	t.Parallel()

	ActiveRooms.Set(5)
	value := testutil.ToFloat64(ActiveRooms)

	if value != 5 {
		t.Errorf("ActiveRooms = %v, want 5", value)
	}
}

func TestActivePlayersIncDec(t *testing.T) {
	t.Parallel()

	initial := testutil.ToFloat64(ActivePlayers)

	ActivePlayers.Inc()
	after := testutil.ToFloat64(ActivePlayers)

	if after != initial+1 {
		t.Errorf("ActivePlayers after Inc() = %v, want %v", after, initial+1)
	}

	ActivePlayers.Dec()
	final := testutil.ToFloat64(ActivePlayers)

	if final != initial {
		t.Errorf("ActivePlayers after Dec() = %v, want %v", final, initial)
	}
}

func TestRoomsCreated(t *testing.T) {
	t.Parallel()

	initial := testutil.ToFloat64(RoomsCreated)

	RoomsCreated.Inc()
	after := testutil.ToFloat64(RoomsCreated)

	if after <= initial {
		t.Errorf("RoomsCreated should increase, got %v", after)
	}
}

func TestRequestDuration(t *testing.T) {
	t.Parallel()

	RequestDuration.WithLabelValues("test_method").Observe(0.5)

	count := testutil.CollectAndCount(RequestDuration)
	if count == 0 {
		t.Error("RequestDuration should have metrics")
	}
}

func TestErrorsTotal(t *testing.T) {
	t.Parallel()

	initial := testutil.ToFloat64(ErrorsTotal.WithLabelValues("test_error"))

	ErrorsTotal.WithLabelValues("test_error").Inc()
	after := testutil.ToFloat64(ErrorsTotal.WithLabelValues("test_error"))

	if after != initial+1 {
		t.Errorf("ErrorsTotal = %v, want %v", after, initial+1)
	}
}

func TestWebSocketConnections(t *testing.T) {
	t.Parallel()

	WebSocketConnections.Set(10)
	value := testutil.ToFloat64(WebSocketConnections)

	if value != 10 {
		t.Errorf("WebSocketConnections = %v, want 10", value)
	}
}

func TestMessagesReceived(t *testing.T) {
	t.Parallel()

	MessagesReceived.WithLabelValues("join").Inc()

	count := testutil.ToFloat64(MessagesReceived.WithLabelValues("join"))
	if count == 0 {
		t.Error("MessagesReceived should have count > 0")
	}
}

func TestGameDuration(t *testing.T) {
	t.Parallel()

	GameDuration.Observe(120.5)

	count := testutil.CollectAndCount(GameDuration)
	if count == 0 {
		t.Error("GameDuration should have metrics")
	}
}

func TestMetricsRegistered(t *testing.T) {
	t.Parallel()

	metrics := []prometheus.Collector{
		ActiveRooms,
		ActivePlayers,
		RoomsCreated,
		RoomsCleaned,
		GameDuration,
		RequestDuration,
		ErrorsTotal,
		WebSocketConnections,
		MessagesReceived,
		MessagesSent,
	}

	for _, m := range metrics {
		if m == nil {
			t.Error("metric should not be nil")
		}
	}
}
