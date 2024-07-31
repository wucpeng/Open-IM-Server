package db

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/utils"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math/rand"
	"strconv"
	"time"
)

const cTag = "tag"
const cSendLog = "send_log"

type Tag struct {
	UserID   string   `bson:"user_id"`
	TagID    string   `bson:"tag_id"`
	TagName  string   `bson:"tag_name"`
	UserList []string `bson:"user_list"`
}

func generateTagID(tagName, userID string) string {
	return utils.Md5(tagName + userID + strconv.Itoa(rand.Int()) + time.Now().String())
}

func (d *DataBases) GetUserTags(userID string) ([]Tag, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cTag)
	var tags []Tag
	cursor, err := c.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return tags, err
	}
	if err = cursor.All(ctx, &tags); err != nil {
		return tags, err
	}
	return tags, nil
}

func (d *DataBases) CreateTag(userID, tagName string, userList []string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cTag)
	tagID := generateTagID(tagName, userID)
	tag := Tag{
		UserID:   userID,
		TagID:    tagID,
		TagName:  tagName,
		UserList: userList,
	}
	_, err := c.InsertOne(ctx, tag)
	return err
}

func (d *DataBases) GetTagByID(userID, tagID string) (Tag, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cTag)
	var tag Tag
	err := c.FindOne(ctx, bson.M{"user_id": userID, "tag_id": tagID}).Decode(&tag)
	return tag, err
}

func (d *DataBases) DeleteTag(userID, tagID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cTag)
	_, err := c.DeleteOne(ctx, bson.M{"user_id": userID, "tag_id": tagID})
	return err
}

func (d *DataBases) SetTag(userID, tagID, newName string, increaseUserIDList []string, reduceUserIDList []string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cTag)
	var tag Tag
	if err := c.FindOne(ctx, bson.M{"tag_id": tagID, "user_id": userID}).Decode(&tag); err != nil {
		return err
	}
	if newName != "" {
		_, err := c.UpdateOne(ctx, bson.M{"user_id": userID, "tag_id": tagID}, bson.M{"$set": bson.M{"tag_name": newName}})
		if err != nil {
			return err
		}
	}
	tag.UserList = append(tag.UserList, increaseUserIDList...)
	tag.UserList = utils.RemoveRepeatedStringInList(tag.UserList)
	for _, v := range reduceUserIDList {
		for i2, v2 := range tag.UserList {
			if v == v2 {
				tag.UserList[i2] = ""
			}
		}
	}
	var newUserList []string
	for _, v := range tag.UserList {
		if v != "" {
			newUserList = append(newUserList, v)
		}
	}
	_, err := c.UpdateOne(ctx, bson.M{"user_id": userID, "tag_id": tagID}, bson.M{"$set": bson.M{"user_list": newUserList}})
	if err != nil {
		return err
	}
	return nil
}

func (d *DataBases) GetUserIDListByTagID(userID, tagID string) ([]string, error) {
	var tag Tag
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cTag)
	_ = c.FindOne(ctx, bson.M{"user_id": userID, "tag_id": tagID}).Decode(&tag)
	return tag.UserList, nil
}

type TagUser struct {
	UserID   string `bson:"user_id"`
	UserName string `bson:"user_name"`
}

type TagSendLog struct {
	UserList         []TagUser `bson:"tag_list"`
	SendID           string    `bson:"send_id"`
	SenderPlatformID int32     `bson:"sender_platform_id"`
	Content          string    `bson:"content"`
	SendTime         int64     `bson:"send_time"`
}

func (d *DataBases) SaveTagSendLog(tagSendLog *TagSendLog) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSendLog)
	_, err := c.InsertOne(ctx, tagSendLog)
	return err
}

func (d *DataBases) GetTagSendLogs(userID string, showNumber, pageNumber int32) ([]TagSendLog, error) {
	var tagSendLogs []TagSendLog
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSendLog)
	findOpts := options.Find().SetLimit(int64(showNumber)).SetSkip(int64(showNumber) * (int64(pageNumber) - 1)).SetSort(bson.M{"send_time": -1})
	cursor, err := c.Find(ctx, bson.M{"send_id": userID}, findOpts)
	if err != nil {
		return tagSendLogs, err
	}
	err = cursor.All(ctx, &tagSendLogs)
	if err != nil {
		return tagSendLogs, err
	}
	return tagSendLogs, nil
}
