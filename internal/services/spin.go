package services

import (
	"errors"
	"math/rand"
	"time"

	"github.com/kaelCoding/toyBE/internal/models"
	"gorm.io/gorm"
)

func SpinWheel(db *gorm.DB) (*models.Reward, error) {
	var allRewards []models.Reward
	if err := db.Find(&allRewards).Error; err != nil {
		return nil, err
	}

	var availableRewards []models.Reward
	var totalProbability float64
	for _, r := range allRewards {
		if r.Quantity > 0 || r.Quantity == -1 {
			availableRewards = append(availableRewards, r)
			totalProbability += r.Probability
		}
	}

	if len(availableRewards) == 0 {
		return nil, errors.New("no rewards available to spin")
	}

	if totalProbability == 0 {
		return nil, errors.New("total probability of available rewards is zero")
	}

	source := rand.NewSource(time.Now().UnixNano())
    rng := rand.New(source)
	
	random := rng.Float64() * totalProbability
	cumulativeProbability := 0.0

	for _, reward := range availableRewards {
		cumulativeProbability += reward.Probability
		if random < cumulativeProbability {
			if reward.Quantity != -1 {
				reward.Quantity--
				if err := db.Save(&reward).Error; err != nil {
					return nil, err
				}
			}
			return &reward, nil
		}
	}

	return nil, errors.New("failed to select a reward, please check probabilities")
}