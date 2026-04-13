package main

import (
	"context"
	"log"
	"time"

	"github.com/codingninja/pob-management/config"
	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"github.com/codingninja/pob-management/pkg/database"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func main() {
	cfg := config.Load()

	db, err := database.NewMongoDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	repo := repository.NewCertificateTypeRepository(db)

	certTypes := []struct {
		name        string
		code        string
		description string
	}{
		{"BOSIET", "BOSIET", "Basic Offshore Safety Induction and Emergency Training"},
		{"HUET", "HUET", "Helicopter Underwater Escape Training"},
		{"Offshore Medical", "OFFSHORE-MED", "Offshore Medical Certificate"},
		{"Scaffolding", "SCAFFOLD", "Scaffolding Certification"},
		{"DPR Offshore Permit", "DPR-PERMIT", "Department of Petroleum Resources Offshore Permit"},
		{"Rigging", "RIGGING", "Rigging and Lifting Certification"},
		{"Crane Operator", "CRANE-OP", "Crane Operator Certification"},
		{"Confined Space Entry", "CONFINED-SPACE", "Confined Space Entry Training"},
		{"Fire Fighting", "FIRE-FIGHT", "Offshore Fire Fighting Training"},
		{"First Aid", "FIRST-AID", "Offshore First Aid Certification"},
	}

	ctx := context.Background()
	now := time.Now()

	for _, ct := range certTypes {
		// Check if already exists
		existing, _ := repo.FindByCode(ctx, ct.code)
		if existing != nil {
			log.Printf("Certificate type already exists: %s", ct.code)
			continue
		}

		certType := &domain.CertificateType{
			ID:          bson.NewObjectID(),
			Name:        ct.name,
			Code:        ct.code,
			Description: ct.description,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Create(ctx, certType)
		if err != nil {
			log.Printf("Failed to create %s: %v", ct.code, err)
		} else {
			log.Printf("Created certificate type: %s", ct.code)
		}
	}

	log.Println("Certificate types seeding completed!")
}
