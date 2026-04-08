package service

import (
	"BackendFramework/internal/database"
	"BackendFramework/internal/middleware"
	"BackendFramework/internal/model"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// Mapping Group ID ke Group Name (Hanya Admin=1 dan User=2)
var groupIDMap = map[uint]string{
	1: "admin",
	2: "user",
}

// Map Group Name ke Group ID
var groupNameMap = map[string]uint{
	"admin": 1,
	"user":  2,
}

// GetAllUsers mengambil semua user
func GetAllUsers() ([]model.UserList, error) {
	var users []model.UserList

	err := database.DbWebkita.
		Table("users").
		Select("username, email, group_id, is_aktif").
		Where("deleted_at IS NULL").
		Scan(&users).Error

	if err != nil {
		middleware.LogError(err, "Query Error: GetAllUsers")
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	return users, nil
}

// GetOneUser mengambil satu user berdasarkan username
func GetOneUser(userId string) (*model.UserList, error) {
	if userId == "" {
		return nil, errors.New("username cannot be empty")
	}

	var user model.UserList

	err := database.DbWebkita.
		Table("users").
		Select("username, email, group_id, is_aktif").
		Where("username = ? AND deleted_at IS NULL", userId).
		Scan(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %s", userId)
		}
		middleware.LogError(err, fmt.Sprintf("Data Scan Error: GetOneUser for username %s", userId))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Cek apakah user ditemukan
	if user.Username == "" {
		return nil, fmt.Errorf("user not found: %s", userId)
	}

	return &user, nil
}

// GetOneUserByEmail mengambil satu user berdasarkan email
func GetOneUserByEmail(userEmail string) (*model.UserList, error) {
	if userEmail == "" {
		return nil, errors.New("email cannot be empty")
	}

	var user model.UserList

	err := database.DbWebkita.
		Table("users").
		Select("username, email, group_id, is_aktif").
		Where("email = ? AND deleted_at IS NULL", userEmail).
		Scan(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found with email: %s", userEmail)
		}
		middleware.LogError(err, fmt.Sprintf("Data Scan Error: GetOneUserByEmail for email %s", userEmail))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Cek apakah user ditemukan
	if user.Email == "" {
		return nil, fmt.Errorf("user not found with email: %s", userEmail)
	}

	return &user, nil
}

// InsertUser menambahkan user baru
func InsertUser(userData *model.UserInput) error {
	if userData == nil {
		return errors.New("user data cannot be nil")
	}

	// Validasi dasar
	if userData.Username == "" {
		return errors.New("username cannot be empty")
	}
	if userData.Email == "" {
		return errors.New("email cannot be empty")
	}
	if userData.Password == "" {
		return errors.New("password cannot be empty")
	}

	// Validasi group_id
	if !IsValidGroupID(userData.GroupID) {
		return model.ErrInvalidGroupID
	}

	// Set default is_aktif jika kosong
	isAktif := userData.IsAktif
	if isAktif == "" {
		isAktif = "Y"
	}

	// Validasi IsAktif
	if isAktif != "Y" && isAktif != "N" {
		return model.ErrInvalidIsAktif
	}

	// Konversi UserInput ke User model
	user := userData.ToUser()

	// Insert menggunakan Create
	err := database.DbWebkita.Create(&user).Error
	if err != nil {
		middleware.LogError(err, "Insert Data Failed: InsertUser")
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// GetCurrentUserProfile mengambil profil user yang sedang login berdasarkan user ID dari JWT
func GetCurrentUserProfile(userID uint) (*model.User, error) {
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	var user model.User

	err := database.DbWebkita.
		Where("id = ? AND deleted_at IS NULL AND is_aktif = ?", userID, "Y").
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found with ID: %d", userID)
		}
		middleware.LogError(err, fmt.Sprintf("Get Current User Profile Failed for ID %d", userID))
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return &user, nil
}

func CreateUser(user *model.User) error {
	if user == nil {
		return errors.New("user data cannot be nil")
	}

	// Validasi dasar
	if user.Username == "" {
		return errors.New("username cannot be empty")
	}
	if user.Email == "" {
		return errors.New("email cannot be empty")
	}
	if user.Password == "" {
		return errors.New("password cannot be empty")
	}

	// Validasi group_id
	if !IsValidGroupID(user.GroupID) {
		return model.ErrInvalidGroupID
	}

	// Set default values
	if user.IsAktif == "" {
		user.IsAktif = "Y"
	}
	if user.GroupID == 0 {
		user.GroupID = 2
	}

	// Validasi IsAktif
	if user.IsAktif != "Y" && user.IsAktif != "N" {
		return model.ErrInvalidIsAktif
	}

	err := database.DbWebkita.Create(user).Error
	if err != nil {
		middleware.LogError(err, "Insert Data Failed: CreateUser")
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateUser memperbarui data user
func UpdateUser(userData *model.UserInput) error {
	if userData == nil {
		return errors.New("user data cannot be nil")
	}

	if userData.Username == "" {
		return errors.New("username cannot be empty")
	}

	// Buat map untuk update
	updates := map[string]interface{}{
		"updated_at": gorm.Expr("NOW()"),
	}

	// Update is_aktif jika tidak kosong
	if userData.IsAktif != "" {
		// Validasi IsAktif
		if userData.IsAktif != "Y" && userData.IsAktif != "N" {
			return model.ErrInvalidIsAktif
		}
		updates["is_aktif"] = userData.IsAktif
	}

	// Update email jika tidak kosong
	if userData.Email != "" {
		updates["email"] = userData.Email
	}

	// Update group_id jika tidak 0
	if userData.GroupID != 0 {
		// Validasi group_id
		if !IsValidGroupID(userData.GroupID) {
			return model.ErrInvalidGroupID
		}
		updates["group_id"] = userData.GroupID
	}

	// Update password hanya jika tidak kosong
	if userData.Password != "" {
		updates["password"] = userData.Password
	}

	result := database.DbWebkita.
		Table("users").
		Where("username = ? AND deleted_at IS NULL", userData.Username).
		Updates(updates)

	if result.Error != nil {
		middleware.LogError(result.Error, "Update Data Failed: UpdateUser")
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found or no changes made")
	}

	return nil
}

// UpdateUserProfile memperbarui profil user (first_name, last_name, phone)
func UpdateUserProfile(username string, updates map[string]interface{}) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}

	if len(updates) == 0 {
		return errors.New("no data to update")
	}

	// Tambahkan updated_at
	updates["updated_at"] = gorm.Expr("NOW()")

	result := database.DbWebkita.
		Table("users").
		Where("username = ? AND deleted_at IS NULL", username).
		Updates(updates)

	if result.Error != nil {
		middleware.LogError(result.Error, "Update Profile Failed: UpdateUserProfile")
		return fmt.Errorf("failed to update user profile: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found or no changes made")
	}

	return nil
}

// DeleteUser menghapus user (soft delete)
func DeleteUser(userId string) error {
	if userId == "" {
		return errors.New("username cannot be empty")
	}

	result := database.DbWebkita.
		Table("users").
		Where("username = ? AND deleted_at IS NULL", userId).
		Update("deleted_at", gorm.Expr("NOW()"))

	if result.Error != nil {
		middleware.LogError(result.Error, "Delete Data Failed: DeleteUser")
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// HardDeleteUser menghapus user secara permanen
func HardDeleteUser(userId string) error {
	if userId == "" {
		return errors.New("username cannot be empty")
	}

	result := database.DbWebkita.
		Table("users").
		Where("username = ?", userId).
		Delete(nil)

	if result.Error != nil {
		middleware.LogError(result.Error, "Hard Delete Data Failed: HardDeleteUser")
		return fmt.Errorf("failed to hard delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// RestoreUser mengembalikan user yang telah di-soft delete
func RestoreUser(userId string) error {
	if userId == "" {
		return errors.New("username cannot be empty")
	}

	result := database.DbWebkita.
		Table("users").
		Where("username = ? AND deleted_at IS NOT NULL", userId).
		Update("deleted_at", nil)

	if result.Error != nil {
		middleware.LogError(result.Error, "Restore Data Failed: RestoreUser")
		return fmt.Errorf("failed to restore user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found or already active")
	}

	return nil
}

// CheckUserExists memeriksa apakah user dengan username atau email sudah ada
func CheckUserExists(username, email string) (bool, error) {
	var count int64

	err := database.DbWebkita.
		Table("users").
		Where("(username = ? OR email = ?) AND deleted_at IS NULL", username, email).
		Count(&count).Error

	if err != nil {
		middleware.LogError(err, "Check User Exists Failed")
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return count > 0, nil
}

// CheckUsernameExists memeriksa apakah username sudah ada
func CheckUsernameExists(username string) (bool, error) {
	var count int64

	err := database.DbWebkita.
		Table("users").
		Where("username = ? AND deleted_at IS NULL", username).
		Count(&count).Error

	if err != nil {
		middleware.LogError(err, "Check Username Exists Failed")
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return count > 0, nil
}

// CheckEmailExists memeriksa apakah email sudah ada
func CheckEmailExists(email string) (bool, error) {
	var count int64

	err := database.DbWebkita.
		Table("users").
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error

	if err != nil {
		middleware.LogError(err, "Check Email Exists Failed")
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

// GetUserWithPassword mengambil user beserta password (untuk autentikasi)
func GetUserWithPassword(username string) (*model.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	var user model.User

	err := database.DbWebkita.
		Where("username = ? AND deleted_at IS NULL AND is_aktif = ?", username, "Y").
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		middleware.LogError(err, fmt.Sprintf("Get User with Password Failed for username %s", username))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserWithPasswordByEmail mengambil user beserta password berdasarkan email (untuk autentikasi)
func GetUserWithPasswordByEmail(email string) (*model.User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	var user model.User

	err := database.DbWebkita.
		Where("email = ? AND deleted_at IS NULL AND is_aktif = ?", email, "Y").
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		middleware.LogError(err, fmt.Sprintf("Get User with Password Failed for email %s", email))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByID mengambil user berdasarkan ID
func GetUserByID(userID uint) (*model.User, error) {
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}

	var user model.User

	err := database.DbWebkita.
		Where("id = ? AND deleted_at IS NULL", userID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found with ID: %d", userID)
		}
		middleware.LogError(err, fmt.Sprintf("Get User by ID Failed for ID %d", userID))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// UpdateUserPassword memperbarui password user
func UpdateUserPassword(username, newPassword string) error {
	if username == "" {
		return errors.New("username cannot be empty")
	}
	if newPassword == "" {
		return errors.New("new password cannot be empty")
	}

	result := database.DbWebkita.
		Table("users").
		Where("username = ? AND deleted_at IS NULL", username).
		Updates(map[string]interface{}{
			"password":   newPassword,
			"updated_at": gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		middleware.LogError(result.Error, "Update Password Failed: UpdateUserPassword")
		return fmt.Errorf("failed to update password: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdateUserPasswordByEmail memperbarui password user berdasarkan email
func UpdateUserPasswordByEmail(email, newPassword string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}
	if newPassword == "" {
		return errors.New("new password cannot be empty")
	}

	result := database.DbWebkita.
		Table("users").
		Where("email = ? AND deleted_at IS NULL", email).
		Updates(map[string]interface{}{
			"password":   newPassword,
			"updated_at": gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		middleware.LogError(result.Error, "Update Password Failed: UpdateUserPasswordByEmail")
		return fmt.Errorf("failed to update password: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// GetAllGroups mengembalikan daftar ID grup yang valid
func GetAllGroups() []uint {
	ids := make([]uint, 0, len(groupIDMap))
	for id := range groupIDMap {
		ids = append(ids, id)
	}
	return ids
}

// GetGroupName mengembalikan nama grup berdasarkan ID
func GetGroupName(groupID uint) (string, bool) {
	name, exists := groupIDMap[groupID]
	return name, exists
}

// GetGroupID mengembalikan ID grup berdasarkan nama
func GetGroupID(groupName string) (uint, bool) {
	id, exists := groupNameMap[groupName]
	return id, exists
}

// IsValidGroupID memeriksa apakah ID grup valid (hanya 1 atau 2)
func IsValidGroupID(groupID uint) bool {
	_, exists := groupIDMap[groupID]
	return exists
}

// IsValidGroupName memeriksa apakah nama grup valid
func IsValidGroupName(groupName string) bool {
	_, exists := groupNameMap[groupName]
	return exists
}
