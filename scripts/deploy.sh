#!/bin/bash
dir=$(pwd)
project_name="wos-redeem-discord-bot"
projectpath="/root/${project_name}/"
logfile="/var/log/${project_name}.log"
if [ ! -f "$logfile" ]; then
    sudo mkdir ${projectpath}
fi
sudo rm -f ${projectpath}/${project_name}
sudo cp "$dir"/${project_name}/${project_name} ${projectpath}/
sudo cp scripts/${project_name}.conf /etc/supervisor/conf.d/
# first time load new conf
if [ ! -f "$logfile" ]; then
    sudo supervisorctl reload
fi
if [ -f "$logfile" ]; then
    sudo rm "$logfile"
fi
sudo supervisorctl update ${project_name}
sudo supervisorctl restart ${project_name}