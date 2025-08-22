package loyalty

import (
	"log"
	"time"

	"github.com/kaelCoding/toyBE/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VIPLevel struct {
    Level                   int
    Threshold               float64
    Discount                float64
    MaintenanceRequirement  float64
}

var VIPLevelMap = map[int]VIPLevel{
    4: {Level: 4, Threshold: 8000000, Discount: 0.07, MaintenanceRequirement: 0},
    3: {Level: 3, Threshold: 5000000, Discount: 0.05, MaintenanceRequirement: 5000000 * 0.3},
    2: {Level: 2, Threshold: 2500000, Discount: 0.03, MaintenanceRequirement: 2500000 * 0.3},
    1: {Level: 1, Threshold: 1000000, Discount: 0.02, MaintenanceRequirement: 1000000 * 0.3},
    0: {Level: 0, Threshold: 0, Discount: 0, MaintenanceRequirement: 0},
}

var VIPLevelsSorted = []VIPLevel{
    VIPLevelMap[4], VIPLevelMap[3], VIPLevelMap[2], VIPLevelMap[1], VIPLevelMap[0],
}

func UpdateUserLoyaltyStatus(tx *gorm.DB, userID uint, orderAmount float64) error {
    var user models.User
    if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
        return err
    }

    user.TotalSpent += orderAmount
    user.MaintenanceSpending += orderAmount

    previousLevel := user.VIPLevel

    for _, level := range VIPLevelsSorted {
        if user.TotalSpent >= level.Threshold && user.VIPLevel < level.Level {
            user.VIPLevel = level.Level
            log.Printf("User ID %d được thăng hạng lên VIP %d", user.ID, user.VIPLevel)
            
            user.MaintenanceSpending = 0
            if level.Level < 4 {
                newExpiryDate := time.Now().AddDate(0, 3, 0)
                user.VIPExpiryDate = &newExpiryDate
            } else {
                user.VIPExpiryDate = nil
            }
            break
        }
    }

    if user.VIPLevel == previousLevel && user.VIPLevel > 0 && user.VIPLevel < 4 {
        currentLevelInfo := GetVIPLevelInfo(user.VIPLevel)
        if user.MaintenanceSpending >= currentLevelInfo.MaintenanceRequirement {
            log.Printf("User ID %d đã duy trì thành công hạng VIP %d", user.ID, user.VIPLevel)
            user.MaintenanceSpending = 0
            newExpiryDate := time.Now().AddDate(0, 3, 0)
            user.VIPExpiryDate = &newExpiryDate
        }
    }
    
    user.DiscountPercentage = GetVIPLevelInfo(user.VIPLevel).Discount
    
    return tx.Save(&user).Error
}

func CheckAndApplyDemotions(db *gorm.DB) {
    log.Println("Bắt đầu chạy tác vụ kiểm tra và hạ cấp VIP...")
    now := time.Now()
    var expiredUsers []models.User

    db.Where("vip_level > 0 AND vip_level < 4 AND vip_expiry_date < ?", now).Find(&expiredUsers)

    if len(expiredUsers) == 0 {
        log.Println("Không có user nào bị hạ cấp.")
        return
    }

    for _, user := range expiredUsers {
        log.Printf("User ID %d (VIP %d) đã hết hạn. Bị hạ cấp xuống VIP %d.", user.ID, user.VIPLevel, user.VIPLevel-1)
        user.VIPLevel--
        
        user.MaintenanceSpending = 0
        if user.VIPLevel > 0 {
            newExpiryDate := time.Now().AddDate(0, 3, 0)
            user.VIPExpiryDate = &newExpiryDate
        } else {
            user.VIPExpiryDate = nil
        }
        
        user.DiscountPercentage = GetVIPLevelInfo(user.VIPLevel).Discount
        
        db.Save(&user)
    }
    log.Printf("Hoàn thành tác vụ hạ cấp cho %d user.", len(expiredUsers))
}

func GetVIPLevelInfo(level int) VIPLevel {
    if l, ok := VIPLevelMap[level]; ok {
        return l
    }
    return VIPLevelMap[0]
}