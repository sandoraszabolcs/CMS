package repository

import (
	"context"
	"testing"

	"github.com/szabolcs/cms/internal/domain"
)

func TestFindOpenCheckin_Integration(t *testing.T) {
	t.Skip("requires postgres — run with: go test -tags=integration ./internal/repository/...")

	// Pattern for integration testing:
	//
	// db := setupTestDB(t) // connect to test postgres, run migrations
	// repo := NewPostgresValidationRepo(db)
	//
	// // Insert a checkin
	// checkin := domain.ValidationEvent{
	//     CardID:    "CMS-TEST",
	//     VehicleID: "BUS-101",
	//     EventType: domain.EventCheckin,
	//     StopID:    "S1",
	//     Lat:       44.4268,
	//     Lng:       26.1025,
	// }
	// inserted, err := repo.InsertEvent(context.Background(), checkin)
	// if err != nil {
	//     t.Fatal(err)
	// }
	//
	// // Should find the open checkin
	// found, err := repo.FindOpenCheckin(context.Background(), "CMS-TEST")
	// if err != nil {
	//     t.Fatal(err)
	// }
	// if found.ID != inserted.ID {
	//     t.Fatalf("expected ID %d, got %d", inserted.ID, found.ID)
	// }
	//
	// // Insert a checkout
	// checkout := domain.ValidationEvent{
	//     CardID:    "CMS-TEST",
	//     VehicleID: "BUS-101",
	//     EventType: domain.EventCheckout,
	//     StopID:    "S2",
	//     Lat:       44.4361,
	//     Lng:       26.1006,
	// }
	// _, err = repo.InsertEvent(context.Background(), checkout)
	// if err != nil {
	//     t.Fatal(err)
	// }
	//
	// // Should no longer find open checkin
	// _, err = repo.FindOpenCheckin(context.Background(), "CMS-TEST")
	// if !errors.Is(err, domain.ErrNotFound) {
	//     t.Fatalf("expected ErrNotFound, got %v", err)
	// }

	_ = context.Background()
	_ = domain.ErrNotFound
}
