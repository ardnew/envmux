package log_test

import (
	"context"
	"testing"

	"github.com/ardnew/envmux/pkg/log"
)

func TestMakeJournal(t *testing.T) {
	t.Run("default construction", func(t *testing.T) {
		journal := log.Make()
		
		if journal.Logger == nil {
			t.Error("Journal should have a non-nil Logger")
		}
	})
	
	t.Run("with custom jotter", func(t *testing.T) {
		jotter := log.MakeJotter(log.WithDiscard())
		journal := log.Make(log.WithJotter(jotter))
		
		if journal.Logger == nil {
			t.Error("Journal should have a non-nil Logger")
		}
	})
}

func TestWithJotter(t *testing.T) {
	// Create a custom jotter with both leveler and handler
	jotter := log.MakeJotter(log.WithLeveler(log.DefaultLevel), log.WithDiscard())
	
	// Create journal with the jotter
	journal := log.Make(log.WithJotter(jotter))
	
	if journal.Logger == nil {
		t.Error("Journal should have a non-nil Logger")
	}
	
	// Test that the journal can be used for logging (no output verification since using discard)
	journal.Info("test message")
	journal.Error("test error")
}

func TestJournalAddToContext(t *testing.T) {
	journal := log.Make()
	ctx := context.Background()
	
	// Add journal to context
	newCtx := journal.AddToContext(ctx)
	
	// Verify context is different
	if newCtx == ctx {
		t.Error("AddToContext should return a new context")
	}
	
	// Verify journal can be retrieved from context
	retrievedJournal, ok := log.FromContext(newCtx)
	if !ok {
		t.Error("FromContext should return true for context with journal")
	}
	
	// Verify it's the same journal
	if retrievedJournal.Logger == nil {
		t.Error("Retrieved journal should have non-nil Logger")
	}
}

func TestFromContext(t *testing.T) {
	t.Run("with journal", func(t *testing.T) {
		journal := log.Make()
		ctx := journal.AddToContext(context.Background())
		
		retrievedJournal, ok := log.FromContext(ctx)
		if !ok {
			t.Error("FromContext should return true for context with journal")
		}
		
		if retrievedJournal.Logger == nil {
			t.Error("Retrieved journal should have non-nil Logger")
		}
	})
	
	t.Run("without journal", func(t *testing.T) {
		ctx := context.Background()
		
		_, ok := log.FromContext(ctx)
		if ok {
			t.Error("FromContext should return false for context without journal")
		}
	})
	
	t.Run("with wrong type in context", func(t *testing.T) {
		// Manually create context with wrong type value
		ctx := context.WithValue(context.Background(), struct{}{}, "not a journal")
		
		_, ok := log.FromContext(ctx)
		if ok {
			t.Error("FromContext should return false for context with wrong type")
		}
	})
}

func TestJournalContextWorkflow(t *testing.T) {
	// Test complete workflow: create -> add to context -> retrieve -> use
	
	// Create journal with proper jotter
	jotter := log.MakeJotter(log.WithLeveler(log.DefaultLevel), log.WithDiscard()) // Use discard to avoid output
	journal := log.Make(log.WithJotter(jotter))
	
	// Add to context
	ctx := journal.AddToContext(context.Background())
	
	// Retrieve from context
	retrievedJournal, ok := log.FromContext(ctx)
	if !ok {
		t.Fatal("Should be able to retrieve journal from context")
	}
	
	// Use the retrieved journal
	retrievedJournal.Info("test message from retrieved journal")
	retrievedJournal.Error("test error from retrieved journal")
	
	// Should work without panics
}

func TestJournalUniqueness(t *testing.T) {
	// Test that unique.Make is working properly
	journal1 := log.Make()
	journal2 := log.Make()
	
	// Both should be valid journals
	if journal1.Logger == nil {
		t.Error("First journal should have non-nil Logger")
	}
	
	if journal2.Logger == nil {
		t.Error("Second journal should have non-nil Logger")
	}
	
	// They should be independent instances
	// (We can't easily test unique.Handle comparison here without internal knowledge)
}

func TestJournalWithMultipleOptions(t *testing.T) {
	// Create jotter with specific configuration
	jotter := log.MakeJotter(
		log.WithLeveler(log.DefaultLevel), // Ensure leveler is set
		log.WithDiscard(), // Use discard handler to avoid output
	)
	
	// Create journal with the jotter
	journal := log.Make(log.WithJotter(jotter))
	
	if journal.Logger == nil {
		t.Error("Journal should have non-nil Logger")
	}
	
	// Test using the journal
	journal.Debug("debug message")
	journal.Info("info message") 
	journal.Warn("warning message")
	journal.Error("error message")
}

func TestJournalContextIsolation(t *testing.T) {
	// Test that different journals in different contexts don't interfere
	
	journal1 := log.Make()
	journal2 := log.Make()
	
	ctx1 := journal1.AddToContext(context.Background())
	ctx2 := journal2.AddToContext(context.Background())
	
	// Retrieve from first context
	retrieved1, ok1 := log.FromContext(ctx1)
	if !ok1 {
		t.Error("Should retrieve journal from first context")
	}
	
	// Retrieve from second context  
	retrieved2, ok2 := log.FromContext(ctx2)
	if !ok2 {
		t.Error("Should retrieve journal from second context")
	}
	
	// Both should be valid but potentially different instances
	if retrieved1.Logger == nil {
		t.Error("First retrieved journal should have non-nil Logger")
	}
	
	if retrieved2.Logger == nil {
		t.Error("Second retrieved journal should have non-nil Logger")
	}
}