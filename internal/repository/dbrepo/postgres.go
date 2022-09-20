package dbrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/rest5387/myApp1/internal/models"
)

// Test function
func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) InsertUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if user.Created_at.IsZero() {
		user.Created_at = time.Now()
		user.Updated_at = time.Now()
	}
	stmt := `INSERT INTO users (first_name, last_name, email, password, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := m.DB.ExecContext(ctx, stmt, user.FirstName, user.LastName, user.Email, user.PasswordHash, user.Created_at, user.Updated_at)
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) SearchUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var user models.User
	query := `SELECT id, first_name, last_name, password, access_level FROM users WHERE email=$1`

	row := m.DB.QueryRowContext(ctx, query, email)

	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.PasswordHash, &user.AccessLevel)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *postgresDBRepo) SearchUserByUID(uid uint) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var user models.User
	query := `SELECT first_name, last_name, email, password, access_level FROM users WHERE id=$1`

	row := m.DB.QueryRowContext(ctx, query, uid)

	err := row.Scan(&user.FirstName, &user.LastName, &user.Email, &user.PasswordHash, &user.AccessLevel)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *postgresDBRepo) InsertPost(post models.Post) (uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if post.Created_at.IsZero() {
		post.Created_at = time.Now()
		post.Updated_at = time.Now()
	}

	stmt := `INSERT INTO posts (uid, likes, content, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5)`
	_, err := m.DB.ExecContext(ctx, stmt, post.UID, post.Likes, post.Content, post.Created_at, post.Updated_at)
	if err != nil {
		return 0, err
	}

	var pid uint
	query := `SELECT id FROM posts WHERE uid=$1 AND updated_at=$2`
	row := m.DB.QueryRowContext(ctx, query, post.UID, post.Updated_at)
	err = row.Scan(&pid)
	if err != nil {
		fmt.Printf("select pid error: %s\n", err.Error())
		return 0, err
	}
	return pid, nil
}

func (m *postgresDBRepo) SearchPostByPID(pid uint) (*models.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var post models.Post
	query := `SELECT uid, likes, content, created_at, updated_at FROM posts WHERE id=$1`

	post.ID = pid
	row := m.DB.QueryRowContext(ctx, query, pid)
	err := row.Scan(&post.UID, &post.Likes, &post.Content, &post.Created_at, &post.Updated_at)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (m *postgresDBRepo) SearchPIDsByUID(uid uint) ([]uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var pidset []uint
	var pid uint
	query := `SELECT id FROM posts WHERE uid=$1 ORDER BY created_at DESC`

	rows, err := m.DB.QueryContext(ctx, query, uid)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		if rows.Scan(&pid) != nil {
			return nil, err
		}
		pidset = append(pidset, pid)
	}

	return pidset, nil
}

func (m *postgresDBRepo) UpdatePostByPID(pid uint, post models.Post) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	post.Updated_at = time.Now()

	stmt := `UPDATE posts SET content=$1, updated_at=$2 WHERE id=$3`
	_, err := m.DB.ExecContext(ctx, stmt, post.Content, post.Updated_at, pid)
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresDBRepo) DeletePostByPID(pid uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	stmt := `DELETE FROM posts WHERE id=$1`
	_, err := m.DB.ExecContext(ctx, stmt, pid)
	if err != nil {
		return err
	}
	return nil
}
