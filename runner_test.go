package runner_test

import (
	"context"
	"testing"
	"time"

	"github.com/maddiesch/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunningCommand(t *testing.T) {
	t.Run("it collects the stdout", func(t *testing.T) {
		out, err := runner.Command("echo", `Testing command output`).Run(context.Background())

		require.NoError(t, err)

		assert.Equal(t, "Testing command output\n", string(out.Stdout))
		assert.Equal(t, "", string(out.Stderr))
	})

	t.Run("given an expiring context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := runner.Command("sleep", "5").Run(ctx)

		assert.Error(t, err)
	})

	t.Run("given an invalid command", func(t *testing.T) {
		_, err := runner.Command("go-runner-test", "foo", "bar").Run(context.Background())

		assert.Error(t, err)
	})
}
