package data

import (
	"time"

	"gorm.io/gorm"
)

type Organization struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `json:"name" gorm:"not null;unique"`

	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

type User struct {
	ID uint `gorm:"primaryKey"`

	Username string `json:"username" gorm:"not null;unique"`

	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

type InstallationUser struct {
	ID uint `gorm:"primaryKey"`

	ExternalID string `json:"external_id" gorm:"not null;unique"`

	InstallationID uint         `json:"installation_id" gorm:"not null"`
	Installation   Installation `gorm:"foreignKey:InstallationID"`

	UserID uint `json:"_user_id" gorm:"not null"`
	User   User `gorm:"foreignKey:UserID"`

	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

type Installation struct {
	ID uint `gorm:"primaryKey"`

	InstallationID string `json:"installation_id" gorm:"not null;unique"`
	Platform       string `json:"platform" gorm:"not null"`

	OrganizationID uint         `json:"organization_id" gorm:"not null"`
	Organization   Organization `gorm:"foreignKey:OrganizationID"`

	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

type Kudos struct {
	ID uint `gorm:"primaryKey"`

	FromUserID uint `json:"from_user_id" gorm:"not null"`
	FromUser   User `gorm:"foreignKey:FromUserID"`

	ToUserID uint `json:"to_user_id" gorm:"not null"`
	ToUser   User `gorm:"foreignKey:ToUserID"`

	Description string `json:"description" gorm:"type:text;not null"`

	InstallationID uint         `json:"installation_id" gorm:"not null"`
	Installation   Installation `gorm:"foreignKey:InstallationID"`

	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

func (db *Database) CreateOrganization(name string) (*Organization, error) {
	organization := Organization{
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tx := db.connection.Create(&organization)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &organization, nil
}

func (db *Database) GetKudusCountForUser(installationID string, username string) (int64, error) {
	var count int64
	tx := db.connection.Model(&Kudos{}).
		Joins("JOIN installation_users ON kudos.from_user_id = installation_users.user_id").
		Joins("JOIN installations ON installation_users.installation_id = installations.id").
		Joins("JOIN users ON installation_users.user_id = users.id").
		Where("installations.installation_id = ? AND users.username = ?", installationID, username).
		Count(&count)

	if tx.Error != nil {
		return 0, tx.Error
	}

	return count, nil
}



func (db *Database) CreateInstallation(platform string, organizationID uint) (*Installation, error) {
	installation := Installation{
		Platform:       platform,
		OrganizationID: organizationID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	tx := db.connection.Create(&installation)

	if tx.Error != nil {
		return nil, tx.Error
	}
	return &installation, nil
}

func createUserIfNotExists(tx *gorm.DB, externalId string, installationID string) (*InstallationUser, error) {
	var installationUser InstallationUser
	tx = tx.Table("installation_users").Joins("left join installations on installations.id = installation_users.installation_id").
		Where("installations.installation_id = ? AND installation_users.external_id = ?", installationID, externalId).First(&installationUser)

	if tx.Error != nil {
		return nil, tx.Error
	}

	if tx.RowsAffected == 0 {
		// Get installation
		var installation Installation
		tx = tx.Where("installation_id = ?", installationID).First(&installation)
		if tx.Error != nil {
			return nil, tx.Error
		}

		// Create User
		user := User{
			Username:  externalId,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		tx = tx.Create(&user)

		if tx.Error != nil {
			return nil, tx.Error
		}

		// Create Installation User
		installationUser = InstallationUser{
			ExternalID:     externalId,
			InstallationID: installation.ID,
			UserID:         user.ID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		tx := tx.Create(&installationUser)
		if tx.Error != nil {
			return nil, tx.Error
		}

	}

	return &installationUser, nil
}

func (db *Database) CreateKudos(fromExternalUsername string, toExternalUsername string, description string, installationID string) (*Kudos, error) {
	// Find From User with ExternalID and InstallationID

	var kudos Kudos

	db.connection.Transaction(func(tx *gorm.DB) error {

		fromInstallationUser, err := createUserIfNotExists(tx, fromExternalUsername, installationID)
		if err != nil {
			return err
		}

		toInstallationUser, err := createUserIfNotExists(tx, toExternalUsername, installationID)
		if err != nil {
			return err
		}

		kudos = Kudos{
			FromUserID:     fromInstallationUser.UserID,
			ToUserID:       toInstallationUser.UserID,
			Description:    description,
			InstallationID: fromInstallationUser.InstallationID,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		tx = tx.Create(&kudos)
		if tx.Error != nil {
			return tx.Error
		}

		return nil
	})

	return &kudos, nil
}
