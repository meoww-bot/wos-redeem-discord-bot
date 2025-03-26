package mongodb

import (
	"context"
	"log"
	"sort"
	"wos-redeem-discord-bot/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	userColl = client.Database("wos").Collection("user")
)

func AddUser(user models.Player) error {

	_, err := userColl.InsertOne(getContext(), user)

	if err != nil {
		if !mongo.IsDuplicateKeyError(err) {
			log.Println(err)
		}

	}

	return err
}

func UpdateUser(user models.Player) error {

	_, err := GetUser(user.FID)

	if err == mongo.ErrNoDocuments {
		return AddUser(user)
	}

	filter := bson.M{"fid": user.FID}

	_, err = userColl.ReplaceOne(context.TODO(), filter, user)

	return err
}

func RemoveUser(UserId int) error {

	filter := bson.M{"fid": UserId}

	_, err := userColl.DeleteOne(context.TODO(), filter)

	return err
}

func GetUser(UserId int) (models.Player, error) {

	var user models.Player

	filter := bson.M{"fid": UserId}

	err := userColl.FindOne(context.TODO(), filter).Decode(&user)

	return user, err

}

func GetAllUser() ([]models.Player, error) {

	var users []models.Player

	cursor, err := userColl.Find(context.Background(), bson.M{})

	if err != nil {
		return users, err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user models.Player
		cursor.Decode(&user)
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].FID < users[j].FID
	})

	return users, nil
}
