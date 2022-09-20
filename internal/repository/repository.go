package repository

import (
	"github.com/rest5387/myApp1/internal/models"
)

type SQLDatabaseRepo interface {
	// User table ops.
	InsertUser(user models.User) error
	SearchUserByEmail(email string) (*models.User, error)
	SearchUserByUID(uid uint) (*models.User, error)
	// Post table ops.
	InsertPost(post models.Post) (uint, error)
	SearchPostByPID(pid uint) (*models.Post, error)
	SearchPIDsByUID(uid uint) ([]uint, error)
	UpdatePostByPID(pid uint, post models.Post) error
	DeletePostByPID(pid uint) error
}

type Neo4jRepo interface {
	InsertUser(models.User) error
	DeleteUser(uid uint) error
	InsertCard(post models.Post) error
	DeleteCard(pid uint) error
	InsertFollow(follower uint, beFollowed uint) error
	DeleteFollow(follower uint, beFollowed uint) error
	SearchFollow(follower uint, beFollowed uint) (bool, error)
}
