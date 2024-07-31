package db

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math/rand"
	"strconv"
	"time"
)

const cWorkMoment = "work_moment"

type WorkMoment struct {
	WorkMomentID         string            `bson:"work_moment_id"`
	UserID               string            `bson:"user_id"`
	UserName             string            `bson:"user_name"`
	FaceURL              string            `bson:"face_url"`
	Content              string            `bson:"content"`
	LikeUserList         []*WorkMomentUser `bson:"like_user_list"`
	AtUserList           []*WorkMomentUser `bson:"at_user_list"`
	PermissionUserList   []*WorkMomentUser `bson:"permission_user_list"`
	Comments             []*Comment        `bson:"comments"`
	PermissionUserIDList []string          `bson:"permission_user_id_list"`
	Permission           int32             `bson:"permission"`
	CreateTime           int32             `bson:"create_time"`
}

type WorkMomentUser struct {
	UserID   string `bson:"user_id"`
	UserName string `bson:"user_name"`
}

type Comment struct {
	UserID        string `bson:"user_id" json:"user_id"`
	UserName      string `bson:"user_name" json:"user_name"`
	ReplyUserID   string `bson:"reply_user_id" json:"reply_user_id"`
	ReplyUserName string `bson:"reply_user_name" json:"reply_user_name"`
	ContentID     string `bson:"content_id" json:"content_id"`
	Content       string `bson:"content" json:"content"`
	CreateTime    int32  `bson:"create_time" json:"create_time"`
}

func generateWorkMomentID(userID string) string {
	return utils.Md5(userID + strconv.Itoa(rand.Int()) + time.Now().String())
}

func generateWorkMomentCommentID(workMomentID string) string {
	return utils.Md5(workMomentID + strconv.Itoa(rand.Int()) + time.Now().String())
}

func (d *DataBases) CreateOneWorkMoment(workMoment *WorkMoment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	workMomentID := generateWorkMomentID(workMoment.UserID)
	workMoment.WorkMomentID = workMomentID
	workMoment.CreateTime = int32(time.Now().Unix())
	_, err := c.InsertOne(ctx, workMoment)
	return err
}

func (d *DataBases) DeleteOneWorkMoment(workMomentID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	_, err := c.DeleteOne(ctx, bson.M{"work_moment_id": workMomentID})
	return err
}

func (d *DataBases) DeleteComment(workMomentID, contentID, opUserID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	_, err := c.UpdateOne(ctx, bson.D{{"work_moment_id", workMomentID},
		{"$or", bson.A{
			bson.D{{"user_id", opUserID}},
			bson.D{{"comments", bson.M{"$elemMatch": bson.M{"user_id": opUserID}}}},
		},
		}}, bson.M{"$pull": bson.M{"comments": bson.M{"content_id": contentID}}})
	return err
}

func (d *DataBases) GetWorkMomentByID(workMomentID string) (*WorkMoment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	workMoment := &WorkMoment{}
	err := c.FindOne(ctx, bson.M{"work_moment_id": workMomentID}).Decode(workMoment)
	return workMoment, err
}

func (d *DataBases) LikeOneWorkMoment(likeUserID, userName, workMomentID string) (*WorkMoment, bool, error) {
	workMoment, err := d.GetWorkMomentByID(workMomentID)
	if err != nil {
		return nil, false, err
	}
	var isAlreadyLike bool
	for i, user := range workMoment.LikeUserList {
		if likeUserID == user.UserID {
			isAlreadyLike = true
			workMoment.LikeUserList = append(workMoment.LikeUserList[0:i], workMoment.LikeUserList[i+1:]...)
		}
	}
	if !isAlreadyLike {
		workMoment.LikeUserList = append(workMoment.LikeUserList, &WorkMomentUser{UserID: likeUserID, UserName: userName})
	}
	log.NewDebug("", utils.GetSelfFuncName(), workMoment)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	_, err = c.UpdateOne(ctx, bson.M{"work_moment_id": workMomentID}, bson.M{"$set": bson.M{"like_user_list": workMoment.LikeUserList}})
	return workMoment, !isAlreadyLike, err
}

func (d *DataBases) SetUserWorkMomentsLevel(userID string, level int32) error {
	return nil
}

func (d *DataBases) CommentOneWorkMoment(comment *Comment, workMomentID string) (WorkMoment, error) {
	comment.ContentID = generateWorkMomentCommentID(workMomentID)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	var workMoment WorkMoment
	err := c.FindOneAndUpdate(ctx, bson.M{"work_moment_id": workMomentID}, bson.M{"$push": bson.M{"comments": comment}}).Decode(&workMoment)
	return workMoment, err
}

func (d *DataBases) GetUserSelfWorkMoments(userID string, showNumber, pageNumber int32) ([]WorkMoment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	var workMomentList []WorkMoment
	findOpts := options.Find().SetLimit(int64(showNumber)).SetSkip(int64(showNumber) * (int64(pageNumber) - 1)).SetSort(bson.M{"create_time": -1})
	result, err := c.Find(ctx, bson.M{"user_id": userID}, findOpts)
	if err != nil {
		return workMomentList, nil
	}
	err = result.All(ctx, &workMomentList)
	return workMomentList, err
}

func (d *DataBases) GetUserWorkMoments(opUserID, userID string, showNumber, pageNumber int32, friendIDList []string) ([]WorkMoment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	var workMomentList []WorkMoment
	findOpts := options.Find().SetLimit(int64(showNumber)).SetSkip(int64(showNumber) * (int64(pageNumber) - 1)).SetSort(bson.M{"create_time": -1})
	result, err := c.Find(ctx, bson.D{ // 等价条件: select * from
		{"user_id", userID},
		{"$or", bson.A{
			bson.D{{"permission", constant.WorkMomentPermissionCantSee}, {"permission_user_id_list", bson.D{{"$nin", bson.A{opUserID}}}}},
			bson.D{{"permission", constant.WorkMomentPermissionCanSee}, {"permission_user_id_list", bson.D{{"$in", bson.A{opUserID}}}}},
			bson.D{{"permission", constant.WorkMomentPublic}},
		}},
	}, findOpts)
	if err != nil {
		return workMomentList, nil
	}
	err = result.All(ctx, &workMomentList)
	return workMomentList, err
}

func (d *DataBases) GetUserFriendWorkMoments(showNumber, pageNumber int32, userID string, friendIDList []string) ([]WorkMoment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	var workMomentList []WorkMoment
	findOpts := options.Find().SetLimit(int64(showNumber)).SetSkip(int64(showNumber) * (int64(pageNumber) - 1)).SetSort(bson.M{"create_time": -1})
	var filter bson.D
	permissionFilter := bson.D{
		{"$or", bson.A{
			bson.D{{"permission", constant.WorkMomentPermissionCantSee}, {"permission_user_id_list", bson.D{{"$nin", bson.A{userID}}}}},
			bson.D{{"permission", constant.WorkMomentPermissionCanSee}, {"permission_user_id_list", bson.D{{"$in", bson.A{userID}}}}},
			bson.D{{"permission", constant.WorkMomentPublic}},
		}}}
	if config.Config.WorkMoment.OnlyFriendCanSee {
		filter = bson.D{
			{"$or", bson.A{
				bson.D{{"user_id", userID}}, //self
				bson.D{{"$and", bson.A{permissionFilter, bson.D{{"user_id", bson.D{{"$in", friendIDList}}}}}}},
			},
			},
		}
	} else {
		filter = bson.D{
			{"$or", bson.A{
				bson.D{{"user_id", userID}}, //self
				permissionFilter,
			},
			},
		}
	}
	result, err := c.Find(ctx, filter, findOpts)
	if err != nil {
		return workMomentList, err
	}
	err = result.All(ctx, &workMomentList)
	return workMomentList, err
}
