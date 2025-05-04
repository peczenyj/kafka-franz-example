package kafkafranzexample_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kslog"

	testcontainers_redpanda "github.com/testcontainers/testcontainers-go/modules/redpanda"
)

func TestMinimal(t *testing.T) {
	const (
		username = "foo"
		password = "bar"
	)
	const redpandaDefaultImage = "docker.redpanda.com/redpandadata/redpanda:v23.3.3"
	redpandaContainer, err := testcontainers_redpanda.Run(t.Context(),
		redpandaDefaultImage,
		testcontainers.WithLogger(log.TestLogger(t)),
		testcontainers_redpanda.WithEnableSASL(),
		testcontainers_redpanda.WithEnableKafkaAuthorization(),
		testcontainers_redpanda.WithNewServiceAccount(username, password),
		testcontainers_redpanda.WithSuperusers(username),
		testcontainers_redpanda.WithEnableSchemaRegistryHTTPBasicAuth(),
	)

	t.Cleanup(func() {
		if terr := testcontainers.TerminateContainer(redpandaContainer); terr != nil {
			t.Logf("failed to terminate container: %s", terr)
		}
	})

	if err != nil {
		t.Fatalf("testcontainers_redpanda.Run(ctx, ...): %v", err)
	}

	state, err := redpandaContainer.State(t.Context())
	if err != nil {
		t.Fatalf("failed to get container state: %v", err)
	}

	if !state.Running {
		t.Fatalf("expecting redpanda container running, not %q", state.Status)
	}

	broker, err := redpandaContainer.KafkaSeedBroker(t.Context())
	if err != nil {
		t.Fatalf("unable to fetch kafka seed broker: %v", err)
	}

	client, err := kgo.NewClient(
		// no SASL to force issue
		kgo.SeedBrokers(broker),
		kgo.WithLogger(kslog.New(slog.Default())),
		kgo.ConsumeTopics("foo"),
		kgo.ConsumerGroup("foo-bar"),
	)
	if err != nil {
		t.Fatalf("unable to create kafka franz client: %v", err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx)

	t.Logf("client pint return error: %v, closing...", err)

	begin := time.Now()

	client.Close()

	took := time.Since(begin)

	t.Logf("closing client after %s", took)
}
