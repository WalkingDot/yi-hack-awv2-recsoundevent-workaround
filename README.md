# Sound Event Recording Workaround for [YI Hack Allwinner-v2](https://github.com/roleoroleo/yi-hack-Allwinner-v2)

My camera doesn't support video recording on sound detection/event. Therefore I wrote this workaround.  

Copy the files into the following locations:  

`/tmp/sd/yi-hack/startup.sh`  
`/tmp/sd/yi-hack/bin/recsoundevent`  

After copying the files or changing arguments, the camera must be restarted. 


Optional arguments can be added into the "startup.sh":  

**-t**  
Recording duration in minutes (1-60, default: 10).  

**-d**  
dB value to trigger recording (30-90, default: Cam dB setting).  
*(With the default cam setting, the sound sensitivity is limited to 50 dB)*  

Examples:  
Record 5 minutes: `/tmp/sd/yi-hack/bin/recsoundevent -t 5 &`  
dB value to 37: `/tmp/sd/yi-hack/bin/recsoundevent -d 37 &`  
Both options: `/tmp/sd/yi-hack/bin/recsoundevent -t 15 -d 30 &`  
