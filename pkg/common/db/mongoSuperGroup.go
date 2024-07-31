package db

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/utils"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"
)

const cSuperGroup = "super_group"
const cUserToSuperGroup = "user_to_super_group"

type SuperGroup struct {
	GroupID string `bson:"group_id" json:"groupID"`
	//MemberNumCount int      `bson:"member_num_count"`
	MemberIDList []string `bson:"member_id_list" json:"memberIDList"`
}

type UserToSuperGroup struct {
	UserID      string   `bson:"user_id" json:"userID"`
	GroupIDList []string `bson:"group_id_list" json:"groupIDList"`
}

func superGroupIndexGen(groupID string, seqSuffix uint32) string {
	return "super_group_" + groupID + ":" + strconv.FormatInt(int64(seqSuffix), 10)
}

func (d *DataBases) CreateSuperGroup(groupID string, initMemberIDList []string, memberNumCount int) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSuperGroup)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	sCtx := mongo.NewSessionContext(ctx, session)
	superGroup := SuperGroup{
		GroupID:      groupID,
		MemberIDList: initMemberIDList,
	}
	_, err = c.InsertOne(sCtx, superGroup)
	if err != nil {
		_ = session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	var users []UserToSuperGroup
	for _, v := range initMemberIDList {
		users = append(users, UserToSuperGroup{
			UserID: v,
		})
	}
	upsert := true
	opts := &options.UpdateOptions{
		Upsert: &upsert,
	}
	c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cUserToSuperGroup)
	//_, err = c.UpdateMany(sCtx, bson.M{"user_id": bson.M{"$in": initMemberIDList}}, bson.M{"$addToSet": bson.M{"group_id_list": groupID}}, opts)
	//if err != nil {
	//	session.AbortTransaction(ctx)
	//	return utils.Wrap(err, "transaction failed")
	//}
	for _, userID := range initMemberIDList {
		_, err = c.UpdateOne(sCtx, bson.M{"user_id": userID}, bson.M{"$addToSet": bson.M{"group_id_list": groupID}}, opts)
		if err != nil {
			_ = session.AbortTransaction(ctx)
			return utils.Wrap(err, "transaction failed")
		}

	}
	return err
}

func (d *DataBases) GetSuperGroup(groupID string) (SuperGroup, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSuperGroup)
	superGroup := SuperGroup{}
	err := c.FindOne(ctx, bson.M{"group_id": groupID}).Decode(&superGroup)
	return superGroup, err
}

func (d *DataBases) AddUserToSuperGroup(groupID string, userIDList []string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSuperGroup)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	sCtx := mongo.NewSessionContext(ctx, session)
	if err != nil {
		return utils.Wrap(err, "start transaction failed")
	}
	_, err = c.UpdateOne(sCtx, bson.M{"group_id": groupID}, bson.M{"$addToSet": bson.M{"member_id_list": bson.M{"$each": userIDList}}})
	if err != nil {
		_ = session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cUserToSuperGroup)
	var users []UserToSuperGroup
	for _, v := range userIDList {
		users = append(users, UserToSuperGroup{
			UserID: v,
		})
	}
	upsert := true
	opts := &options.UpdateOptions{
		Upsert: &upsert,
	}
	for _, userID := range userIDList {
		_, err = c.UpdateOne(sCtx, bson.M{"user_id": userID}, bson.M{"$addToSet": bson.M{"group_id_list": groupID}}, opts)
		if err != nil {
			_ = session.AbortTransaction(ctx)
			return utils.Wrap(err, "transaction failed")
		}
	}
	_ = session.CommitTransaction(ctx)
	return err
}

func (d *DataBases) RemoverUserFromSuperGroup(groupID string, userIDList []string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSuperGroup)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	sCtx := mongo.NewSessionContext(ctx, session)
	_, err = c.UpdateOne(ctx, bson.M{"group_id": groupID}, bson.M{"$pull": bson.M{"member_id_list": bson.M{"$in": userIDList}}})
	if err != nil {
		_ = session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	err = d.RemoveGroupFromUser(ctx, sCtx, groupID, userIDList)
	if err != nil {
		_ = session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	_ = session.CommitTransaction(ctx)
	return err
}

func (d *DataBases) GetSuperGroupByUserID(userID string) (UserToSuperGroup, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cUserToSuperGroup)
	var user UserToSuperGroup
	_ = c.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
	return user, nil
}

func (d *DataBases) DeleteSuperGroup(groupID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSuperGroup)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	sCtx := mongo.NewSessionContext(ctx, session)
	superGroup := &SuperGroup{}
	result := c.FindOneAndDelete(sCtx, bson.M{"group_id": groupID})
	err = result.Decode(superGroup)
	if err != nil {
		session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	if err = d.RemoveGroupFromUser(ctx, sCtx, groupID, superGroup.MemberIDList); err != nil {
		session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	session.CommitTransaction(ctx)
	return nil
}

func (d *DataBases) RemoveGroupFromUser(ctx, sCtx context.Context, groupID string, userIDList []string) error {
	var users []UserToSuperGroup
	for _, v := range userIDList {
		users = append(users, UserToSuperGroup{
			UserID: v,
		})
	}
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cUserToSuperGroup)
	_, err := c.UpdateOne(sCtx, bson.M{"user_id": bson.M{"$in": userIDList}}, bson.M{"$pull": bson.M{"group_id_list": groupID}})
	if err != nil {
		return utils.Wrap(err, "UpdateOne transaction failed")
	}
	return err
}
