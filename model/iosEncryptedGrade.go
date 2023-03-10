package model

import (
	log "github.com/sirupsen/logrus"
	"test-student-lecture-selection-algorithm/ios_crypto"
)

// IOSEncryptedGrade is a grade that can be encrypted.
// Whether it is currently encrypted or not is indicated by the IsEncrypted field.
type IOSEncryptedGrade struct {
	ID           uint      `gorm:"primaryKey"`
	Device       IOSDevice `gorm:"constraint:OnDelete:CASCADE"`
	DeviceID     string    `gorm:"index;not null"`
	LectureTitle string    `gorm:"not null"`
	Grade        string    `gorm:"not null"`
	IsEncrypted  bool      `gorm:"not null,default:true"`
}

func (e *IOSEncryptedGrade) Encrypt(key string) error {
	encryptedTitle, err := ios_crypto.SymmetricEncrypt(e.LectureTitle, key)
	if err != nil {
		log.WithError(err).Error("Failed to encrypt lecture title")
		return err
	}

	encryptedGrade, err := ios_crypto.SymmetricEncrypt(e.Grade, key)
	if err != nil {
		log.WithError(err).Error("Failed to encrypt grade")
		return err
	}

	e.LectureTitle = encryptedTitle.String()
	e.Grade = encryptedGrade.String()
	e.IsEncrypted = true

	return nil
}

func (e *IOSEncryptedGrade) Decrypt(key string) error {
	decryptedTitle, err := ios_crypto.SymmetricDecrypt(ios_crypto.EncryptedString(e.LectureTitle), key)

	if err != nil {
		return err
	}

	decryptedGrade, err := ios_crypto.SymmetricDecrypt(ios_crypto.EncryptedString(e.Grade), key)

	if err != nil {
		return err
	}

	e.LectureTitle = decryptedTitle
	e.Grade = decryptedGrade
	e.IsEncrypted = false

	return nil
}
