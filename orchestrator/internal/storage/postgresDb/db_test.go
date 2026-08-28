package postgresDb

import (
	"context"
	"testing"
)

func TestDbDriver(t *testing.T) {
	dbURL := "postgres://ZORO:SanjiIsBetterThanZoro@localhost:5432/DAGOrchestrator" //?sslmode=disable
	ctx := context.Background()

	// Note: Returning (error, value) is against Go conventions,
	// but kept as-is to match your existing method signatures.
	dbDriver, err := CreatePostgresDriver(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to create a Db: %v", err)
	}

	// Optional but recommended: defer closing your DB connection
	defer dbDriver.pool.Close()

	workflowID, err := dbDriver.CreateWorkflow(ctx)
	if err != nil {
		t.Fatalf("failed to create a workflow: %v", err)
	}
	t.Logf("workflow id: %d", workflowID)

	// The expected number of ready tasks for each evaluation iteration
	expectedReadyCounts := []int{3, 2, 1, 1}

	for i, expectedCount := range expectedReadyCounts {
		iteration := i + 1

		if err != nil {
			t.Fatalf("iteration %d: failed to evaluate the task table: %v", iteration, err)
		}

		tasksInfo, err := dbDriver.TesttempGettingInfo(ctx, workflowID)
		if err != nil {
			t.Fatalf("iteration %d: failed to get task info: %v", iteration, err)
		}

		for _, info := range tasksInfo {
			t.Logf("Iteration %d - Type: %s, Status: %s", iteration, info.Task_type, info.Task_status)
		}

		t.Logf("end of iteration %d", iteration)

		readyTasks, err := dbDriver.Evaluate(ctx, workflowID)
		if err != nil {
			t.Fatalf("iteration %d: failed to fetch ready tasks: %v", iteration, err)
		}

		// Fixed the copy-paste error in the error message here
		if len(readyTasks) != expectedCount {
			t.Errorf("iteration %d: expected ready tasks=%d got %d", iteration, expectedCount, len(readyTasks))
		}

		// Update tasks to "COMPLETED" (Skip this on the final iteration to match your original logic)
		if i < len(expectedReadyCounts)-1 {
			for _, id := range readyTasks {
				err = dbDriver.UpdateTask(ctx, id, workflowID, "COMPLETED")
				if err != nil {
					t.Fatalf("iteration %d: failed to update task %v: %v", iteration, id, err)
				}
			}
		}
	}
}
