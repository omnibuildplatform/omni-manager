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

type BuildLog struct {
	Id            int       `gorm:"primaryKey"`
	Arch          string    ` description:"architecture"`
	Release       string    ` description:"release openEuler Version"`
	BuildType     string    ` description:"iso , zip ...."`
	BasePkg       string    ` description:"default package"`
	CustomPkg     string    ` description:"custom"`
	UserId        int       ` description:"user id"`
	UserName      string    ` description:"user name"`
	CreateTime    time.Time ` description:"create time"`
	Status        string    ` description:"current status :running ,success, failed"`
	JobName       string    ` description:"pod name"`
	DownloadUrl   string    ` description:"download the result of build iso file"`
	ConfigMapName string    ` description:"configMap name"`
}

func (t *BuildLog) TableName() string {
	return "build_log"
}

func (t *BuildLog) ToString() string {
	return fmt.Sprintf("id:%d;Architecture:%s;EulerVersion:%s;OutFormat:%s;UserId:%d;UserName:%s;JobName:%s", t.Id, t.Arch, t.Release, t.BuildType, t.UserId, t.UserName, t.JobName)
}

// AddBuildLog insert a new ImageMeta into database and returns
// last inserted Id on success.
func AddBuildLog(m *BuildLog) (id int64, err error) {
	o := util.GetDB()
	result := o.Create(m)
	return int64(m.Id), result.Error
}

// GetBuildLogById retrieves ImageMeta by Id. Returns error if
// Id doesn't exist
func GetBuildLogById(id int) (v *BuildLog, err error) {
	o := util.GetDB()
	v = &BuildLog{Id: id}
	o.First(v, id)
	return v, err
}

// GetAllBuildLog retrieves all ImageMeta matches certain condition. Returns empty list if
// no records exist
func GetAllBuildLog(query map[string]string, fields []string, sortby []string, order []string,
	offset int64, limit int64) (ml []interface{}, err error) {
	return nil, err
}

// GetMyBuildLogs query my build history
func GetMyBuildLogs(userid int, offset int, limit int) (ml []*BuildLog, err error) {
	o := util.GetDB()
	m := new(BuildLog)
	m.UserId = userid
	ml = make([]*BuildLog, limit)
	sql := fmt.Sprintf("select * from %s where user_id = %d order by id desc limit %d,%d", m.TableName(), userid, offset, limit)
	o.Raw(sql).Scan(&ml)
	return ml, nil
}

// UpdateBuildLogById updates ImageMeta by Id and returns error if
// the record to be updated doesn't exist
func UpdateBuildLogById(m *BuildLog) (err error) {
	o := util.GetDB()
	result := o.Model(m).Updates(m)
	return result.Error
}

// CreateTables
func CreateTables() (err error) {
	o := util.GetDB()
	if !o.Migrator().HasTable(&BuildLog{}) {
		// create table if not exist
		err = o.Migrator().CreateTable(&BuildLog{})
	}

	return
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

	omniImager := `omni-imager --package-list /conf/totalrpms.json --config-file /conf/conf.yaml --build-type ` + buildtype + ` --output-file ` + outputName + ` && curl -vvv -Ffile=@/opt/omni-workspace/` + outputName + ` -Fproject=` + release + `  -FfileType=image '` + util.GetConfig().K8sConfig.FfileType + `'`
	jobInterface := clientset.BatchV1().Jobs(util.GetConfig().K8sConfig.Namespace)
	var backOffLimit int32 = 0
	var tTLSecondsAfterFinished int32 = 1800
	var privileged bool = true
	var ownerReferenceController bool = true
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
					},
				},
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: &tTLSecondsAfterFinished,
		},
	}
	// cmBytes, _ := json.Marshal(jobYaml)
	// fmt.Println("------------cm :", string(cmBytes))
	job, err = jobInterface.Create(context.TODO(), jobYaml, metav1.CreateOptions{})

	return
}
func CheckPodStatus(jobname string) (err error) {
	jobAPI := GetClientSet().BatchV1()
	var job *batchv1.Job
	job, err = jobAPI.Jobs(util.GetConfig().K8sConfig.Namespace).Get(context.TODO(), jobname, metav1.GetOptions{})
	if err != nil {
		return
	}
	jobRedisBytes, redisErr := util.GetJsonByte(fmt.Sprintf("build_log:%s", jobname))
	if redisErr != nil {
		return redisErr
	}

	buildLog := new(BuildLog)
	err = json.Unmarshal(jobRedisBytes, buildLog)
	if err != nil {
		return
	}
	completions := job.Spec.Completions
	backoffLimit := job.Spec.BackoffLimit
	result := make(map[string]interface{})
	result["name"] = jobname
	result["startTime"] = job.Status.StartTime
	// check status
	if job.Status.Succeeded >= *completions {
		result["status"] = JOB_STATUS_SUCCEED
		result["completionTime"] = job.Status.CompletionTime
		if buildLog != nil {
			result["url"] = buildLog.DownloadUrl
			buildLog.Status = JOB_STATUS_SUCCEED
			AddBuildLog(buildLog)
		}

		job = nil
	} else if job.Status.Failed >= *backoffLimit {
		result["status"] = JOB_STATUS_FAILED
		result["error"] = job.Status.String()
		result["completionTime"] = job.Status.CompletionTime
		if buildLog != nil {
			buildLog.Status = JOB_STATUS_FAILED
			AddBuildLog(buildLog)
		}
	} else if job.Status.Succeeded == 0 || job.Status.Failed == 0 {
		result["status"] = JOB_STATUS_RUNNING
	}
	return
}
