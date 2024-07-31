#!/usr/bin/env bash
# 0 open_im_api
# 1 open_im_user
# 2 open_im_friend
# 3 open_im_group
# 4 open_im_auth
# 5 open_im_conversation
# 6 open_im_cache
# 7 open_im_cron_task
# 8 msg_gateway
# 9 msg_transfer
# 10 msg
# 11 push


source ./style_info.cfg
source ./path_info.cfg
source ./function.sh

bin_dir="../bin"
logs_dir="../logs"
sdk_db_dir="../db/sdk/"
#Automatically created when there is no bin, logs folder
if [ ! -d $bin_dir ]; then
  mkdir -p $bin_dir
fi
if [ ! -d $logs_dir ]; then
  mkdir -p $logs_dir
fi
if [ ! -d $sdk_db_dir ]; then
  mkdir -p $sdk_db_dir
fi

#begin path
begin_path=$PWD

i=$1

cd $begin_path
service_path=${service_source_root[$i]}
cd $service_path
make install
if [ $? -ne 0 ]; then
      echo -e "${RED_PREFIX}${service_names[$i]} build failed ${COLOR_SUFFIX}\n"
      exit -1
      else
       echo -e "${GREEN_PREFIX}${service_names[$i]} successfully be built ${COLOR_SUFFIX}\n"
fi

echo -e ${YELLOW_PREFIX}"all services build success"${COLOR_SUFFIX}
