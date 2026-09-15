package download

import (
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func TestCurrentLayer(t *testing.T) {
	sizes := []int64{10, 20, 30}
	if currentLayer(0, sizes) != 1 {
		t.Fatal("start at layer 1")
	}
	if currentLayer(15, sizes) != 2 {
		t.Fatal("15 should be layer 2")
	}
	if currentLayer(100, sizes) != 3 {
		t.Fatal("complete should be last layer")
	}
	if currentLayer(5, nil) != 0 {
		t.Fatal("no layers")
	}
}

func TestAggregatorPercentageAndETA(t *testing.T) {
	agg := NewAggregator("t1", []int64{50, 50})
	agg.lastComplete = 0
	agg.lastTime = time.Now().Add(-time.Second)
	ev := agg.EventFromUpdate(v1.Update{Complete: 50, Total: 100})
	if ev.Percentage < 49 || ev.Percentage > 51 {
		t.Fatalf("percentage %v", ev.Percentage)
	}
	if ev.DownloadedBytes != 50 {
		t.Fatalf("downloaded %d", ev.DownloadedBytes)
	}
	if ev.CurrentLayer != 1 && ev.CurrentLayer != 2 {
		t.Fatalf("layer %d", ev.CurrentLayer)
	}
	if ev.TotalLayers != 2 {
		t.Fatalf("total layers %d", ev.TotalLayers)
	}
	if ev.SpeedBytes <= 0 {
		t.Fatalf("expected speed, got %d", ev.SpeedBytes)
	}
	if ev.ETASeconds <= 0 {
		t.Fatalf("expected eta, got %d", ev.ETASeconds)
	}
}

func TestAggregatorCacheFields(t *testing.T) {
	agg := NewAggregator("t1", []int64{10, 10})
	agg.SetCache(1, 1)
	ev := agg.Snapshot(10, 20)
	if ev.CachedLayers != 1 || ev.MissingLayers != 1 {
		t.Fatalf("%+v", ev)
	}
}
