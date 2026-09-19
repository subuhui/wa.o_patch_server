package logic

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"shorebird-server/internal/db"
)

func (l *AppLogic) RenameApp(appID, name string) error {
	if strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 128 {
		return fmt.Errorf("name must contain 1 to 128 characters and not be blank")
	}
	var app db.App
	if err := l.svcCtx.DB.First(&app, "id = ?", appID).Error; err != nil {
		return err
	}
	return l.svcCtx.DB.Model(&app).Update("display_name", name).Error
}

func (l *AppLogic) DeleteChannel(appID, channelID string) error {
	return l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var channel db.Channel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND app_id = ?", channelID, appID).First(&channel).Error; err != nil {
			return err
		}
		switch channel.Name {
		case "stable", "beta", "staging":
			return fmt.Errorf("built-in channels cannot be deleted")
		}
		// Keep patch history but remove its association with the deleted channel.
		if err := tx.Model(&db.Patch{}).Where("channel_id = ?", channel.ID).Update("channel_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&channel).Error
	})
}
