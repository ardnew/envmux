package log_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/ardnew/envmux/pkg/log"
)

func TestMakeJotter(t *testing.T) {
	t.Run("default construction", func(t *testing.T) {
		jotter := log.MakeJotter()
		
		// Should have default level and handler
		if jotter.Level() != log.DefaultLevel {
			t.Errorf("Expected default level %v, got %v", log.DefaultLevel, jotter.Level())
		}
		
		// Should be enabled for info level by default
		if !jotter.Enabled(context.Background(), slog.LevelInfo) {
			t.Error("Jotter should be enabled for info level by default")
		}
	})
	
	t.Run("with custom leveler", func(t *testing.T) {
		jotter := log.MakeJotter(log.WithLeveler(slog.LevelDebug))
		
		if jotter.Level() != slog.LevelDebug {
			t.Errorf("Expected debug level, got %v", jotter.Level())
		}
		
		if !jotter.Enabled(context.Background(), slog.LevelDebug) {
			t.Error("Jotter should be enabled for debug level")
		}
	})
}

func TestWithLeveler(t *testing.T) {
	tests := []struct {
		name     string
		level    slog.Level
		testLevel slog.Level
		enabled  bool
	}{
		{"debug level", slog.LevelDebug, slog.LevelDebug, true},
		{"debug level for info", slog.LevelDebug, slog.LevelInfo, true},
		{"info level for debug", slog.LevelInfo, slog.LevelDebug, false},
		{"warn level", slog.LevelWarn, slog.LevelWarn, true},
		{"error level", slog.LevelError, slog.LevelError, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jotter := log.MakeJotter(log.WithLeveler(tt.level))
			
			if jotter.Level() != tt.level {
				t.Errorf("Expected level %v, got %v", tt.level, jotter.Level())
			}
			
			enabled := jotter.Enabled(context.Background(), tt.testLevel)
			if enabled != tt.enabled {
				t.Errorf("Expected enabled=%v for level %v, got %v", tt.enabled, tt.testLevel, enabled)
			}
		})
	}
}

func TestWithHandler(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	
	jotter := log.MakeJotter(log.WithHandler(handler))
	
	// Test that the handler is used
	record := slog.Record{}
	record.Message = "test message"
	record.Level = slog.LevelInfo
	
	err := jotter.Handle(context.Background(), record)
	if err != nil {
		t.Errorf("Handle failed: %v", err)
	}
	
	if !strings.Contains(buf.String(), "test message") {
		t.Error("Handler should have processed the message")
	}
}

func TestWithText(t *testing.T) {
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	
	jotter := log.MakeJotter(log.WithText(&buf, opts))
	
	// Test that it creates a text handler
	record := slog.Record{}
	record.Message = "text test"
	record.Level = slog.LevelDebug
	
	err := jotter.Handle(context.Background(), record)
	if err != nil {
		t.Errorf("Handle failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "text test") {
		t.Error("Text handler should have processed the message")
	}
	
	// Text handler output should be human-readable (not JSON)
	if strings.Contains(output, `{"`) {
		t.Error("Text handler should not produce JSON output")
	}
}

func TestWithJSON(t *testing.T) {
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	
	jotter := log.MakeJotter(log.WithJSON(&buf, opts))
	
	// Test that it creates a JSON handler
	record := slog.Record{}
	record.Message = "json test"
	record.Level = slog.LevelDebug
	
	err := jotter.Handle(context.Background(), record)
	if err != nil {
		t.Errorf("Handle failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "json test") {
		t.Error("JSON handler should have processed the message")
	}
	
	// JSON handler output should contain JSON structure
	if !strings.Contains(output, `"msg"`) {
		t.Error("JSON handler should produce JSON output")
	}
}

func TestWithDiscard(t *testing.T) {
	jotter := log.MakeJotter(log.WithDiscard())
	
	// Test that messages are discarded
	record := slog.Record{}
	record.Message = "should be discarded"
	record.Level = slog.LevelError
	
	err := jotter.Handle(context.Background(), record)
	if err != nil {
		t.Errorf("Handle failed: %v", err)
	}
	
	// Since it's a discard handler, there's no output to verify
	// Just ensure no error occurred
}

func TestJotterEnabled(t *testing.T) {
	jotter := log.MakeJotter(log.WithLeveler(slog.LevelWarn))
	
	tests := []struct {
		level   slog.Level
		enabled bool
	}{
		{slog.LevelDebug, false},
		{slog.LevelInfo, false},
		{slog.LevelWarn, true},
		{slog.LevelError, true},
	}
	
	for _, tt := range tests {
		enabled := jotter.Enabled(context.Background(), tt.level)
		if enabled != tt.enabled {
			t.Errorf("Enabled(%v) = %v, want %v", tt.level, enabled, tt.enabled)
		}
	}
}

func TestJotterWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	jotter := log.MakeJotter(log.WithText(&buf, nil))
	
	// Add attributes
	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}
	
	handlerWithAttrs := jotter.WithAttrs(attrs)
	
	// Test that the handler with attributes works
	record := slog.Record{}
	record.Message = "test with attrs"
	record.Level = slog.LevelInfo
	
	err := handlerWithAttrs.Handle(context.Background(), record)
	if err != nil {
		t.Errorf("Handle failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "key1=value1") {
		t.Error("Handler should include added attributes")
	}
	if !strings.Contains(output, "key2=42") {
		t.Error("Handler should include added attributes")
	}
}

func TestJotterWithGroup(t *testing.T) {
	var buf bytes.Buffer
	jotter := log.MakeJotter(log.WithText(&buf, nil))
	
	// Add group
	handlerWithGroup := jotter.WithGroup("testgroup")
	
	// Test that the handler with group works
	record := slog.Record{}
	record.Message = "test with group"
	record.Level = slog.LevelInfo
	record.AddAttrs(slog.String("attr", "value"))
	
	err := handlerWithGroup.Handle(context.Background(), record)
	if err != nil {
		t.Errorf("Handle failed: %v", err)
	}
	
	output := buf.String()
	// Group handling varies by handler implementation
	// Just verify no error occurred and message is present
	if !strings.Contains(output, "test with group") {
		t.Error("Handler should include the message")
	}
}

func TestJotterPreventSelfReference(t *testing.T) {
	// Test that WithLeveler prevents self-referential chains
	jotter1 := log.MakeJotter()
	jotter2 := log.MakeJotter(log.WithLeveler(jotter1))
	
	// jotter2 should be jotter1, not a new jotter with jotter1 as leveler
	if jotter2.Level() != jotter1.Level() {
		t.Error("WithLeveler should prevent self-referential chains")
	}
	
	// Test that WithHandler prevents self-referential chains
	jotter3 := log.MakeJotter(log.WithHandler(jotter1))
	
	// jotter3 should be jotter1, not a new jotter with jotter1 as handler  
	if jotter3.Level() != jotter1.Level() {
		t.Error("WithHandler should prevent self-referential chains")
	}
}