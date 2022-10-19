package dbrepo

import (
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/rest5387/myApp1/goapp/internal/models"
)

// Test Function
func (m *neo4jRepo) AllUsers() bool {
	session := m.Neo4j.NewSession(neo4j.SessionConfig{})
	// m.Neo4j.NewSession(neo4j.SessionConfig{})
	session.Close()
	return true
}

// Add the person node and its properties into Neo4j DB.
func (m *neo4jRepo) InsertUser(user models.User) error {
	session := m.Neo4j.NewSession(neo4j.SessionConfig{})
	// if err != nil {
	// 	// fmt.Errorf("new session: %s", err.Error())
	// 	// return err
	// 	return fmt.Errorf("new session: %s", err.Error())
	// }
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run("CREATE (n:Person {uid: $uid, firstName:$firstName, lastName:$lastName})", map[string]interface{}{
			"uid":       user.ID,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
		})
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})
	if err != nil {
		// return err
		return fmt.Errorf("writeTransaction: %s", err.Error())
	}
	return nil
}

// Delete the person node from Neo4j DB.
func (m *neo4jRepo) DeleteUser(uid int) error {
	session := m.Neo4j.NewSession(neo4j.SessionConfig{})
	// if err != nil {
	// 	return err
	// }
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run("MATCH (n:Person {uid: $uid}) "+"DETACH DELETE n", map[string]interface{}{
			"uid": uid,
		})
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})
	if err != nil {
		// return err
		return fmt.Errorf("writeTransaction: %s", err.Error())
	}
	return nil
}

// Add the card node with info and its relationships with writer.
func (m *neo4jRepo) InsertCard(post models.Post) error {
	session := m.Neo4j.NewSession(neo4j.SessionConfig{})
	// if err != nil {
	// 	return err
	// }
	defer session.Close()
	neo4j.DateOf(post.Created_at)
	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(
			"CREATE (n:Card {pid:$pid, content:$content, created_at:$created_at, updated_at:$updated_at}) "+
				"WITH n "+
				"MATCH (writer:Person) WHERE writer.uid = $uid "+
				"CREATE (writer)-[r:WROTE]->(n)",
			map[string]interface{}{
				"pid":        post.ID,
				"content":    post.Content,
				"created_at": neo4j.DateOf(post.Created_at),
				"updated_at": neo4j.DateOf(post.Updated_at),
				"uid":        post.UID,
			})
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})
	if err != nil {
		return err
	}
	return nil
}

// Delete the card node with info and its relationships with writer.
func (m *neo4jRepo) DeleteCard(pid int) error {
	session := m.Neo4j.NewSession(neo4j.SessionConfig{})
	// if err != nil {
	// 	return err
	// }
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run("MATCH (n:Card {pid: $pid}) "+"DETACH DELETE n", map[string]interface{}{
			"pid": pid,
		})
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})
	if err != nil {
		return err
	}
	return nil
}

// Add a follow relationship from follower to beFollowed person.
func (m *neo4jRepo) InsertFollow(follower int, beFollowed int) error {
	session := m.Neo4j.NewSession(neo4j.SessionConfig{})
	// if err != nil {
	// 	return err
	// }
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(
			"MATCH (follower:Person), (beFollowed:Person) "+
				"WHERE follower.uid = $uid1 AND beFollowed.uid = $uid2 "+
				"CREATE (follower)-[:FOLLOW]->(beFollowed)",
			map[string]interface{}{
				"uid1": follower,
				"uid2": beFollowed,
			})
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})
	if err != nil {
		return err
	}
	return nil
}

// Delete a follow relationship from follower to beFollowed person.
func (m *neo4jRepo) DeleteFollow(follower int, beFollowed int) error {
	session := m.Neo4j.NewSession(neo4j.SessionConfig{})
	// if err != nil {
	// 	return err
	// }
	defer session.Close()

	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run(
			"MATCH (follower:Person {uid:$uid1})-[r:FOLLOW]->(beFollowed:Person {uid:$uid2}) "+
				"DELETE r",
			map[string]interface{}{
				"uid1": follower,
				"uid2": beFollowed,
			})
		if err != nil {
			return nil, err
		}
		return result.Consume()
	})
	if err != nil {
		return err
	}
	return nil
}

// Search a follow relationship from follower to beFollowed person.
func (m *neo4jRepo) SearchFollow(follower int, beFollowed int) (bool, error) {
	session := m.Neo4j.NewSession(neo4j.SessionConfig{})
	// if err != nil {
	// 	return false, err
	// }
	defer session.Close()

	followed, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var followed bool
		result, err := tx.Run(
			"MATCH (follower:Person {uid:$uid1})-[r:FOLLOW]->(beFollowed:Person {uid:$uid2}) "+
				"RETURN COUNT(r)",
			map[string]interface{}{
				"uid1": follower,
				"uid2": beFollowed,
			})
		if err != nil {
			return nil, err
		}
		for result.Next() {
			// followed = result.Record().Values()[0].(int64) > 0
			followed = result.Record().Values[0].(int64) > 0
		}
		return followed, nil
	})
	if err != nil {
		return false, err
	}
	if !followed.(bool) {
		return false, nil
	}
	return true, nil
}

// GetAllFollowedUID search all uids that user followed.
func (m *neo4jRepo) GetAllFollowedUID(uid int) ([]int, error) {
	session := m.Neo4j.NewSession(neo4j.SessionConfig{})
	// if err != nil {
	// 	return nil, err
	// }
	defer session.Close()

	follows, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		var follows []int
		result, err := tx.Run("MATCH (follower:Person {uid:$uid})-[r:FOLLOW]->(beFollowed:Person) "+
			"return beFollowed.uid",
			map[string]interface{}{
				"uid": uid,
			})

		if err != nil {
			return nil, err
		}
		for result.Next() {
			// uid := result.Record().Values()[0].(int64)
			uid := result.Record().Values[0].(int64)
			follows = append(follows, int(uid))
		}
		return follows, nil
	})

	if err != nil {
		return nil, err
	}
	return follows.([]int), nil
}
