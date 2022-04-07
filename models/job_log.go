package models

import (
	"context"
	"encoding/json"
	"fmt"
	"omni-manager/util"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/yaml.v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//post this body to backend
type BuildParam struct {
	Id        int      `gorm:"primaryKey"`
	Arch      string   ` description:"architecture"`
	Release   string   ` description:"release openEuler Version"`
	BuildType string   ` description:"iso , zip ...."`
	CustomPkg []string ` description:"custom"`
}
type JobLog struct {
	JobName       string    ` description:"pod name" gorm:"primaryKey"`
	Arch          string    ` description:"architecture"`
	Release       string    ` description:"release openEuler Version"`
	BuildType     string    ` description:"iso , zip ...."`
	BasePkg       string    ` gorm:"size:5055"  description:"default package"`
	CustomPkg     string    ` gorm:"size:5055" description:"custom"`
	UserId        int       ` description:"user id"`
	UserName      string    ` description:"user name"`
	CreateTime    time.Time ` description:"create time"`
	Status        string    ` description:"current status :running ,success, failed"`
	DownloadUrl   string    ` description:"download the result of build iso file"`
	ConfigMapName string    ` description:"configMap name"`
}

func (t *JobLog) TableName() string {
	return "job_log"
}

func (t *JobLog) ToString() string {
	return fmt.Sprintf(" Architecture:%s;EulerVersion:%s;OutFormat:%s;UserId:%d;UserName:%s;JobName:%s", t.Arch, t.Release, t.BuildType, t.UserId, t.UserName, t.JobName)
}

// AddJobLog insert a new ImageMeta into database and returns
// last inserted Id on success.
func AddJobLog(m *JobLog) (err error) {
	o := util.GetDB()
	result := o.FirstOrCreate(m)
	return result.Error
}

func GetJobLogByJobName(jobname string) (v *JobLog, err error) {
	o := util.GetDB()
	v = new(JobLog)
	sql := fmt.Sprintf("select * from %s where job_name = '%s' order by create_time desc limit 1", v.TableName(), jobname)
	tx := o.Debug().Raw(sql).Scan(v)
	return v, tx.Error
}

// GetAllJobLog retrieves all ImageMeta matches certain condition. Returns empty list if
// no records exist
func GetAllJobLog(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	return nil, err
}

// GetMyJobLogs query my build history
func GetMyJobLogs(userid int, offset int, limit int) (ml []*JobLog, err error) {
	o := util.GetDB()
	m := new(JobLog)
	m.UserId = userid
	ml = make([]*JobLog, limit)
	sql := fmt.Sprintf("select * from %s where user_id = %d order by create_time desc limit %d,%d", m.TableName(), userid, offset, limit)
	o.Raw(sql).Scan(&ml)
	return ml, nil
}

// UpdateJobLogById updates ImageMeta by Id and returns error if
// the record to be updated doesn't exist
func UpdateJobLogById(m *JobLog) (err error) {
	o := util.GetDB()
	result := o.Model(m).Updates(m)
	return result.Error
}

// UpdateJobLogStatusById
func UpdateJobLogStatusById(jobname, newStatus string) (err error) {
	o := util.GetDB()
	m := new(JobLog)
	sql := fmt.Sprintf("update %s set status='%s' where job_name = '%s'", m.TableName(), newStatus, jobname)
	result := o.Model(m).Exec(sql)
	return result.Error
}

// CreateTables
func CreateTables() (err error) {
	o := util.GetDB()
	if !o.Migrator().HasTable(&JobLog{}) {
		err = o.Migrator().CreateTable(&JobLog{})
	}

	return
}

//Persistence a Job_log  from redis to db
func PersistenceJob(m *JobLog) (err error) {
	err = util.DelKey(CreateRedisJobName(m.JobName), nil)
	if err != nil {
		return
	}
	return AddJobLog(m)

}

//make ConfigMap
func MakeConfigMap(release string, customRpms []string) (cm *v1.ConfigMap) {
	totalPkgs := make(map[string][]string)
	totalPkgs["packages"] = append(util.GetConfig().DefaultPkgList.Packages, customRpms...)
	confYmalConentBytes, err := json.Marshal(totalPkgs)
	if err != nil {
		return
	}

	configMapName := fmt.Sprintf("cmname%d", time.Now().UnixMicro())
	tempdata := make(map[string]string)
	tempdata["working_dir"] = "/opt/omni-workspace"
	tempdata["debug"] = "True"
	tempdata["user_name"] = "root"
	tempdata["user_passwd"] = "openEuler"
	tempdata["installer_configs"] = "/etc/omni-imager/installer_assets/calamares-configs"
	tempdata["systemd_configs"] = "/etc/omni-imager/installer_assets/systemd-configs"
	tempdata["init_script"] = "/etc/omni-imager/init"
	tempdata["installer_script"] = "/etc/omni-imager/runinstaller"
	tempdata["repo_file"] = fmt.Sprintf("/etc/omni-imager/repos/%s.repo", release)
	tempdataBytes, _ := yaml.Marshal(tempdata)

	configMapType := metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "ConfigMap",
	}
	var configImage *v1.ConfigMap
	configImage = &v1.ConfigMap{
		TypeMeta: configMapType,
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: util.GetConfig().K8sConfig.Namespace,
		},
		Data: map[string]string{
			"conf.yaml":      string(tempdataBytes),
			"totalrpms.json": string(confYmalConentBytes),
		},
	}
	cm, err = clientset.CoreV1().ConfigMaps(util.GetConfig().K8sConfig.Namespace).Create(context.TODO(), configImage, metav1.CreateOptions{
		TypeMeta: configImage.TypeMeta,
	})
	if err != nil {
		return
	}
	cm.TypeMeta = configMapType
	cm.Name = configMapName

	return
}

//make job yaml and start job
func MakeJob(cm *v1.ConfigMap, buildtype, release string) (job *batchv1.Job, outputName string, err error) {
	controllerID := uuid.NewV4().String()
	var jobName = fmt.Sprintf(`omni-image-%s`, controllerID)
	outputName = fmt.Sprintf(`openEuler-%s.iso`, controllerID)
	clientset, err := kubernetes.NewForConfig(GetK8sConfig())
	if err != nil {
		return
	}
	// cacheCurl := `curl -vvv -Ffile=@/opt/rootfs_cache/rootfs.tar.gz  -FfileType=image '` + util.GetConfig().K8sConfig.FfileType + `'`
	omniImager := `omni-imager --package-list /conf/totalrpms.json --config-file /conf/conf.yaml --build-type ` + buildtype + ` --output-file ` + outputName + ` && curl -vvv -Ffile=@/opt/omni-workspace/` + outputName + ` -Fproject=` + release + `  -FfileType=image '` + util.GetConfig().K8sConfig.FfileType + `'`

	// omniImager := `omni-imager --package-list /conf/totalrpms.json --config-file /conf/conf.yaml --build-type ` + buildtype + ` --output-file ` + outputName + ` && curl -vvv -Ffile=@/opt/omni-workspace/` + outputName + ` -Fproject=` + release + `  -FfileType=image '` + util.GetConfig().K8sConfig.FfileType + `'`
	jobInterface := clientset.BatchV1().Jobs(util.GetConfig().K8sConfig.Namespace)
	var backOffLimit int32 = 0
	var tTLSecondsAfterFinished int32 = 1800
	var privileged bool = true
	var ownerReferenceController bool = false
	var BlockOwnerDeletion bool = true
	jobYaml := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: util.GetConfig().K8sConfig.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         cm.APIVersion,
					Kind:               cm.Kind,
					Name:               cm.Name,
					Controller:         &ownerReferenceController,
					UID:                cm.UID,
					BlockOwnerDeletion: &BlockOwnerDeletion,
				},
			},
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  jobName,
							Image: util.GetConfig().K8sConfig.Image,
							SecurityContext: &corev1.SecurityContext{
								Privileged: &privileged,
							},
							Command: []string{
								"/bin/sh",
								"-c",
								omniImager,
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "confyaml",
									MountPath: "/conf",
								},
								{
									Name:      "pvcdata",
									MountPath: "/opt/omni-backup",
								},
								{
									Name:      "rootfs",
									MountPath: "/opt/rootfs_cache",
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
					Volumes: []v1.Volume{
						{
							Name: "confyaml",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: cm.Name,
									},
								},
							},
						},
						{
							Name: "pvcdata",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "cce-obs-omni-manager-backend",
								},
							},
						},
						{
							Name: "rootfs",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "cce-sfs-rootfs",
								},
							},
						},
					},
				},
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: &tTLSecondsAfterFinished,
		},
	}
	job, err = jobInterface.Create(context.TODO(), jobYaml, metav1.CreateOptions{})

	return
}
func CreateRedisJobName(jobname string) string {
	return fmt.Sprintf("build_log:%s", jobname)
}
func CheckPodStatus(ns, jobname string) (result map[string]interface{}, job *batchv1.Job, err error) {
	jobAPI := GetClientSet().BatchV1()
	// var job *batchv1.Job
	job, err = jobAPI.Jobs(ns).Get(context.TODO(), jobname, metav1.GetOptions{})
	if err != nil {
		util.Log.Errorf("CheckPodStatus Error:%s", err)
		return
	}

	var JobLog *JobLog
	JobLog, err = GetJobLogByJobName(jobname)
	completions := job.Spec.Completions
	backoffLimit := job.Spec.BackoffLimit
	result = make(map[string]interface{})
	result["name"] = jobname
	result["startTime"] = job.Status.StartTime
	// check status
	if job.Status.Succeeded >= *completions {
		result["status"] = JOB_STATUS_SUCCEED
		result["completionTime"] = job.Status.CompletionTime
		if JobLog != nil {
			result["url"] = JobLog.DownloadUrl
			UpdateJobLogStatusById(jobname, JOB_STATUS_SUCCEED)
		}
		job = nil
	} else if job.Status.Failed > *backoffLimit {
		result["status"] = JOB_STATUS_FAILED
		result["error"] = job.Status.String()
		result["completionTime"] = job.Status.CompletionTime
		UpdateJobLogStatusById(jobname, JOB_STATUS_FAILED)
	} else if job.Status.Succeeded == 0 || job.Status.Failed == 0 {
		result["status"] = JOB_STATUS_RUNNING
	}
	return
}
