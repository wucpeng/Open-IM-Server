package db

import (
	"errors"
	"open_im_sdk/pkg/db/model_struct"
	"open_im_sdk/pkg/utils"
)

func (d *DataBase) GetLoginUser() (*model_struct.LocalUser, error) {
	d.mRWMutex.RLock()
	defer d.mRWMutex.RUnlock()
	var user model_struct.LocalUser
	return &user, utils.Wrap(d.conn.First(&user).Error, "GetLoginUserInfo failed")
}

func (d *DataBase) UpdateLoginUser(user *model_struct.LocalUser) error {
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	t := d.conn.Updates(user)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UpdateLoginUser failed")
}
func (d *DataBase) UpdateLoginUserByMap(user *model_struct.LocalUser, args map[string]interface{}) error {
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	t := d.conn.Model(&user).Updates(args)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UpdateColumnsConversation failed")
}
func (d *DataBase) InsertLoginUser(user *model_struct.LocalUser) error {
	d.mRWMutex.Lock()
	defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Create(user).Error, "InsertLoginUser failed")
}
