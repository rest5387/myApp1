package repository

import (
	"github.com/rest5387/myApp1/goapp/internal/models"
)

type SQLDatabaseRepo interface {
	// User table ops.
	InsertUser(user models.User) error
	SearchUserByEmail(email string) (*models.User, error)
	SearchUserByUID(uid int) (*models.User, error)
	// Post table ops.
	InsertPost(post models.Post) (int, error)
	SearchPostByPID(pid int) (*models.Post, error)
	SearchPIDsByUID(uid int) ([]int, error)
	UpdatePostByPID(pid int, post models.Post) error
	DeletePostByPID(pid int) error

	GetFollowsPIDS(uids []int) ([]int, error)
}

type Neo4jRepo interface {
	InsertUser(models.User) error
	DeleteUser(uid int) error
	InsertCard(post models.Post) error
	DeleteCard(pid int) error
	InsertFollow(follower int, beFollowed int) error
	DeleteFollow(follower int, beFollowed int) error
	SearchFollow(follower int, beFollowed int) (bool, error)
	GetAllFollowedUID(uid int) ([]int, error)
}

type RedisRepo interface {
	Set(key string, value interface{}) error
	Get(key string, dest interface{}) error
	Exists(key string) bool
}
