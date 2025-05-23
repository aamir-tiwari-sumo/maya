/*
Copyright 2018 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pool

import (
	"fmt"
	"strings"
	"time"

	"github.com/aamir-tiwari-sumo/maya/pkg/alertlog"

	apis "github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/v1alpha1"
	zpool "github.com/aamir-tiwari-sumo/maya/pkg/apis/openebs.io/zpool/v1alpha1"
	"github.com/aamir-tiwari-sumo/maya/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

var (
	poolTypeCommand  = map[string]string{"mirrored": "mirror", "raidz": "raidz", "raidz2": "raidz2"}
	defaultGroupSize = map[string]int{"striped": 1, "mirrored": 2, "raidz": 3, "raidz2": 6}
)

// PoolOperator is the name of the tool that makes pool-related operations.
const (
	StatusNoPoolsAvailable = "no pools available"
	ZpoolStatusDegraded    = "DEGRADED"
	ZpoolStatusFaulted     = "FAULTED"
	ZpoolStatusOffline     = "OFFLINE"
	ZpoolStatusOnline      = "ONLINE"
	ZpoolStatusRemoved     = "REMOVED"
	ZpoolStatusUnavail     = "UNAVAIL"
)

//PoolAddEventHandled is a flag representing if the pool has been initially imported or created
var PoolAddEventHandled = false

// PoolNamePrefix is a typed string to store pool name prefix
type PoolNamePrefix string

// ImportedCStorPools is a map of imported cstor pools API config identified via their UID
var ImportedCStorPools map[string]*apis.CStorPool

// CStorZpools is a map of imported cstor pools config at backend identified via their UID
var CStorZpools map[string]zpool.Topology

// PoolPrefix is prefix for pool name
const (
	PoolPrefix PoolNamePrefix = "cstor-"
)

// ImportOptions contains the options to build import command
type ImportOptions struct {
	// CachefileFlag option to use cachefile for import
	CachefileFlag bool

	// DevPath is directory where pool devices resides
	DevPath string

	// dontImport, being true, makes sure `zpool import` command is built
	// without pool name.
	// This way, we know the existence of pool without importing pool
	dontImport bool
}

// RunnerVar the runner variable for executing binaries.
var RunnerVar util.Runner

// ImportPool imports cStor pool if already present.
func ImportPool(cStorPool *apis.CStorPool, importOptions *ImportOptions) (string, error) {
	importAttr := importPoolBuilder(cStorPool, importOptions)

	stdoutStderr, err2 := RunnerVar.RunCommandWithLog(zpool.PoolOperator, importAttr...)
	if err2 != nil {
		klog.Errorf("Unable to import pool: %v, %v devpath: %v cacheflag: %v importAttr: %v",
			err2.Error(), string(stdoutStderr), importOptions.DevPath,
			importOptions.CachefileFlag, importAttr)
		alertlog.Logger.Errorw("",
			"eventcode", "cstor.pool.import.failure",
			"msg", "Failed to import CStor pool",
			"rname", cStorPool.Name,
		)
		return string(stdoutStderr), err2
	}

	klog.Infof("Import command successful with %v %v dontimport: %v importattr: %v out: %v",
		importOptions.CachefileFlag, importOptions.DevPath, importOptions.dontImport,
		importAttr, string(stdoutStderr))

	poolName := string(PoolPrefix) + string(cStorPool.ObjectMeta.UID)
	statusPoolStr := []string{"status", poolName}
	stdoutStderr1, err1 := RunnerVar.RunCombinedOutput(zpool.PoolOperator, statusPoolStr...)
	if err1 != nil {
		klog.Errorf("Unable to get pool status: %v %v", string(stdoutStderr1), statusPoolStr)
		return "", err1
	}
	klog.Infof("pool status: %v", string(stdoutStderr1))
	alertlog.Logger.Infow("",
		"eventcode", "cstor.pool.import.success",
		"msg", "CStor pool imported successfully",
		"rname", cStorPool.Name,
	)
	return string(stdoutStderr), nil
}

// Important learning:

// Below builder will build options something like:
// "import" "-c" "/tmp/pool1.cache" "-o" "cachefile=/tmp/pool1.cache"

// if this line
// `importAttr = append(importAttr, "-o", "cachefile="+cStorPool.Spec.PoolSpec.CacheFile)`
// is changed to
// `importAttr = append(importAttr, "-o cachefile=", cStorPool.Spec.PoolSpec.CacheFile)`

// options will be built like:
// "-o cachefile=" "/tmp/pool1.cache"

// With above thing, `getopt` of zpool treats it as:
// - ` cachefile` as parameter which is wrong (look at <space> at the start)
// - empty value for it, as after '=' nothing is there
// - `/tmp/pool1.cache` as pool name

// Conclusion:
// - append every word, i.e., the one with spaces around it
// - don't append if there are no spaces around it

// importPoolBuilder is to build pool import command.
func importPoolBuilder(cStorPool *apis.CStorPool, importOptions *ImportOptions) []string {
	// populate pool import attributes.
	var importAttr []string
	var cachefileFlag bool
	var devPath string

	// import takes either cachefile or devPath, not both.
	// here, prioritizing devPath. Using zero values for cachefileFlag and devPath
	if importOptions.DevPath != "" {
		devPath = importOptions.DevPath
	} else {
		cachefileFlag = importOptions.CachefileFlag
	}

	importAttr = append(importAttr, "import")
	if cStorPool.Spec.PoolSpec.CacheFile != "" && cachefileFlag {
		importAttr = append(importAttr, "-c", cStorPool.Spec.PoolSpec.CacheFile)
	}

	importAttr = append(importAttr, "-o", "cachefile="+cStorPool.Spec.PoolSpec.CacheFile)

	if devPath != "" {
		importAttr = append(importAttr, "-d", devPath)
	}

	// if dontImport is set to false, build `zpool import` with poolname so that pool gets imported
	if importOptions.dontImport == false {
		importAttr = append(importAttr, string(PoolPrefix)+string(cStorPool.ObjectMeta.UID))
	}
	return importAttr
}

// GetDevPathIfNotSlashDev gets the path from given deviceID if its not prefix
// to "/dev"
func GetDevPathIfNotSlashDev(devID string) string {
	if len(devID) == 0 {
		return ""
	}

	if strings.HasPrefix(devID, "/dev") {
		return ""
	}
	lastindex := strings.LastIndexByte(devID, '/')
	if lastindex == -1 {
		return ""
	}
	devidbytes := []rune(devID)
	return string(devidbytes[0:lastindex])
}

// checkForPoolExistence checks cStor pool existence.
func checkForPoolExistence(cStorPool *apis.CStorPool, blockDeviceList []string) bool {
	var importOptions ImportOptions

	// First device in the list is picked under assumption that all pool devices resides
	// in same place.
	importOptions.DevPath = GetDevPathIfNotSlashDev(blockDeviceList[0])
	importOptions.dontImport = true
	stdoutStderr, _ := ImportPool(cStorPool, &importOptions)
	klog.Infof("checkForPoolExistence output: %v", stdoutStderr)
	return strings.Contains(stdoutStderr, string(PoolPrefix)+string(cStorPool.ObjectMeta.UID))
}

// checkIfPresent is to check if search string is present in array of string.
func checkIfPresent(arrStr []string, searchStr string) bool {
	for _, str := range arrStr {
		if str == searchStr {
			return true
		}
	}
	return false
}

// CreatePool creates a new cStor pool.
func CreatePool(cStorPool *apis.CStorPool, blockDeviceList []string) error {
	// check if pool already imported
	existingPool, err := GetPoolName()
	if err != nil {
		return errors.Errorf("Unable to get poolname %s", err.Error())
	}
	if checkIfPresent(existingPool, string(PoolPrefix)+string(cStorPool.GetUID())) {
		klog.Infof("Pool %s already imported", string(cStorPool.GetUID()))
		return nil
	}

	// check if pool exists but didn't got imported
	exists := checkForPoolExistence(cStorPool, blockDeviceList)
	if exists {
		klog.Errorf("pool %v exists, but failed to import", string(cStorPool.ObjectMeta.UID))
		return errors.Errorf("pool %v exists, but failed to import", string(cStorPool.ObjectMeta.UID))
	}

	err = LabelClear(blockDeviceList)
	if err != nil {
		klog.Errorf(err.Error(), "label clear failed %v", cStorPool.GetUID())
	} else {
		klog.Infof("Label clear successful: %v", string(cStorPool.GetUID()))
	}

	createAttr := createPoolBuilder(cStorPool, blockDeviceList)
	klog.V(4).Info("createAttr : ", createAttr)

	stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, createAttr...)
	if err != nil {
		klog.Errorf("Unable to create pool: %v", string(stdoutStderr))
		alertlog.Logger.Errorw("",
			"eventcode", "cstor.pool.create.failure",
			"msg", "Failed to create CStor pool",
			"rname", cStorPool.Name,
		)
		return errors.Wrapf(err, "zpool create command failed error: %s", string(stdoutStderr))
	}
	alertlog.Logger.Infow("",
		"eventcode", "cstor.pool.create.success",
		"msg", "CStor pool created successfully",
		"rname", cStorPool.Name,
	)
	return nil
}

// createPoolBuilder is to build create pool command.
func createPoolBuilder(cStorPool *apis.CStorPool, blockDeviceList []string) []string {
	// populate pool creation attributes.
	var createAttr []string
	// When block devices of other file formats, say ext4, are used to create cstorpool,
	// it errors out with normal zpool create saying some active file system exists on disk.
	createAttr = append(createAttr, "create")
	if cStorPool.Spec.PoolSpec.CacheFile != "" {
		cachefile := "cachefile=" + cStorPool.Spec.PoolSpec.CacheFile
		createAttr = append(createAttr, "-o", cachefile)
	}

	openebsPoolname := "io.openebs:poolname=" + cStorPool.Name
	createAttr = append(createAttr, "-O", openebsPoolname)

	poolNameUID := string(PoolPrefix) + string(cStorPool.ObjectMeta.UID)
	createAttr = append(createAttr, poolNameUID)
	poolType := cStorPool.Spec.PoolSpec.PoolType
	if poolType == "striped" {
		createAttr = append(createAttr, blockDeviceList...)
		return createAttr
	}
	// To generate pool of the following types:
	// mirrored (grouped by multiples of 2): mirror blockdevice1 blockdevice2 mirror blockdevice3 blockdevice4
	// raidz (grouped by multiples of 3): raidz blockdevice1 blockdevice2 blockdevice3 raidz blockdevice 4 blockdevice5 blockdevice6
	// raidz2 (grouped by multiples of 6): raidz2 blockdevice1 blockdevice2 blockdevice3 blockdevice4 blockdevice5 blockdevice6
	for i, bd := range blockDeviceList {
		if i%defaultGroupSize[poolType] == 0 {
			createAttr = append(createAttr, poolTypeCommand[poolType])
		}
		createAttr = append(createAttr, bd)
	}

	return createAttr
}

// ValidatePool checks for validity of CStorPool resource.
func ValidatePool(cStorPool *apis.CStorPool, devID []string) error {
	poolUID := cStorPool.ObjectMeta.UID
	if len(poolUID) == 0 {
		return fmt.Errorf("Poolname/UID cannot be empty")
	}
	diskCount := len(devID)
	poolType := cStorPool.Spec.PoolSpec.PoolType
	if diskCount < defaultGroupSize[poolType] {
		return errors.Errorf(
			"csp validation failed: expected {%d} blockdevices got {%d}, for pool type {%s}",
			defaultGroupSize[poolType],
			diskCount,
			poolType,
		)
	}
	if diskCount%defaultGroupSize[poolType] != 0 {
		return errors.Errorf(
			"csp validation failed: expected multiples of {%d} blockdevices required got {%d}, for pool type {%s}",
			defaultGroupSize[poolType],
			diskCount,
			poolType,
		)
	}
	return nil
}

// GetPoolName return the pool already created.
func GetPoolName() ([]string, error) {
	GetPoolStr := []string{"get", "-Hp", "name", "-o", "name"}
	poolNameByte, err := RunnerVar.RunStdoutPipe(zpool.PoolOperator, GetPoolStr...)
	if err != nil || len(string(poolNameByte)) == 0 {
		return []string{}, err
	}
	noisyPoolName := string(poolNameByte)
	sepNoisyPoolName := strings.Split(noisyPoolName, "\n")
	var poolNames []string
	for _, poolName := range sepNoisyPoolName {
		poolName = strings.TrimSpace(poolName)
		poolNames = append(poolNames, poolName)
	}
	return poolNames, nil
}

// DeletePool destroys the pool created.
func DeletePool(poolName string) error {
	deletePoolStr := []string{"destroy", poolName}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, deletePoolStr...)
	if err != nil {
		// If pool is missing then do not return error
		if strings.Contains(string(stdoutStderr), "no such pool") {
			klog.Infof("Assuming pool deletion successful for error: %v", string(stdoutStderr))
			return nil
		}
		klog.Errorf("Unable to delete pool : %v", string(stdoutStderr))
		alertlog.Logger.Errorw("",
			"eventcode", "cstor.pool.delete.failure",
			"msg", "Failed to delete CStor pool",
			"rname", poolName,
		)
		return errors.Wrapf(err, "failed to delete pool.. %s", string(stdoutStderr))
	}
	alertlog.Logger.Infow("",
		"eventcode", "cstor.pool.delete.success",
		"msg", "CStor pool deleted successfully",
		"rname", poolName,
	)
	return nil
}

// Capacity finds the capacity of the pool.
// The ouptut of command executed is as follows:
/*
root@cstor-sparse-pool-o8bw-6869f69cc8-jhs6c:/# zpool get size,free,allocated cstor-2ebe403a-f2e2-11e8-87fd-42010a800087
NAME                                        PROPERTY   VALUE  SOURCE
cstor-2ebe403a-f2e2-11e8-87fd-42010a800087  size       9.94G  -
cstor-2ebe403a-f2e2-11e8-87fd-42010a800087  free       9.94G  -
cstor-2ebe403a-f2e2-11e8-87fd-42010a800087  allocated  202K   -
*/
func Capacity(poolName string) (*apis.CStorPoolCapacityAttr, error) {
	capacityPoolStr := []string{"get", "size,free,allocated", poolName}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, capacityPoolStr...)
	if err != nil {
		klog.Errorf("Unable to get pool capacity: %v", string(stdoutStderr))
		return nil, err
	}
	poolCapacity := capacityOutputParser(string(stdoutStderr))
	if strings.TrimSpace(poolCapacity.Used) == "" || strings.TrimSpace(poolCapacity.Free) == "" {
		return nil, fmt.Errorf("Unable to get pool capacity from capacity parser")
	}
	return poolCapacity, nil
}

// PoolStatus finds the status of the pool.
// The ouptut of command(`zpool get -Hp  -ovalue health,io.openebs:readonly <pool-name>`) executed is as follows:

/*
root@cstor-pool-1dvj-854db8dc56-prblp:/# zpool get -Hp  -ovalue health,io.openebs:readonly  cstor-3cbec7b9-578d-11ea-b66e-42010a9a0080
ONLINE
off
*/
// The output is then parsed by poolStatusOutputParser function to get the status of the pool
func Status(poolName string) (string, bool, error) {
	var poolStatus string
	var readOnly bool

	statusPoolStr := []string{"get", "-Hp", "-ovalue", "health,io.openebs:readonly", poolName}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, statusPoolStr...)
	if err != nil {
		klog.Errorf("Unable to get pool status: %v", string(stdoutStderr))
		return "", readOnly, err
	}
	readOnly, poolStatus = poolStatusOutputParser(string(stdoutStderr))

	poolStatus = func(s string) string {
		switch s {
		case ZpoolStatusDegraded:
			return string(apis.CStorPoolStatusDegraded)
		case ZpoolStatusFaulted:
			return string(apis.CStorPoolStatusOffline)
		case ZpoolStatusOffline:
			return string(apis.CStorPoolStatusOffline)
		case ZpoolStatusOnline:
			return string(apis.CStorPoolStatusOnline)
		case ZpoolStatusRemoved:
			return string(apis.CStorPoolStatusDegraded)
		case ZpoolStatusUnavail:
			return string(apis.CStorPoolStatusError)
		default:
			return string(apis.CStorPoolStatusError)
		}
	}(poolStatus)

	return poolStatus, readOnly, nil
}

// poolStatusOutputParser parse output of `zpool status` command to extract the status of the pool.
// ToDo: Need to find some better way e.g contract for zpool command outputs.
func poolStatusOutputParser(output string) (bool, string) {
	var outputStr []string
	var poolStatus string
	var readOnly bool

	outputStr = strings.Split(string(output), "\n")

	if len(outputStr) != 3 {
		klog.Errorf("Invalid input='%s' for poolStatusOutputParser", output)
		return readOnly, poolStatus
	}

	poolStatus = strings.TrimSpace(string(outputStr[0]))
	if outputStr[1] == "on" {
		readOnly = true
	}
	return readOnly, poolStatus
}

// capacityOutputParser parse output of `zpool get` command to extract the capacity of the pool.
// ToDo: Need to find some better way e.g contract for zpool command outputs.
func capacityOutputParser(output string) *apis.CStorPoolCapacityAttr {
	var outputStr []string
	// Initialize capacity object.
	capacity := &apis.CStorPoolCapacityAttr{
		"",
		"",
		"",
	}
	if strings.TrimSpace(string(output)) != "" {
		outputStr = strings.Split(string(output), "\n")
		if !(len(outputStr) < 4) {
			poolCapacityArrTotal := strings.Fields(outputStr[1])
			poolCapacityArrFree := strings.Fields(outputStr[2])
			poolCapacityArrAlloc := strings.Fields(outputStr[3])
			if !(len(poolCapacityArrTotal) < 4 || len(poolCapacityArrFree) < 4) || len(poolCapacityArrAlloc) < 4 {
				capacity.Total = strings.TrimSpace(poolCapacityArrTotal[2])
				capacity.Free = strings.TrimSpace(poolCapacityArrFree[2])
				capacity.Used = strings.TrimSpace(poolCapacityArrAlloc[2])
			}
		}
	}
	return capacity
}

// SetCachefile is to set the cachefile for pool.
func SetCachefile(cStorPool *apis.CStorPool) error {
	poolNameUID := string(PoolPrefix) + string(cStorPool.ObjectMeta.UID)
	setCachefileStr := []string{"set", "cachefile=" + cStorPool.Spec.PoolSpec.CacheFile,
		poolNameUID}
	stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, setCachefileStr...)
	if err != nil {
		klog.Errorf("Unable to set cachefile: %v", string(stdoutStderr))
		return err
	}
	return nil
}

// CheckForZreplInitial is blocking call for checking status of zrepl in cstor-pool container.
func CheckForZreplInitial(ZreplRetryInterval time.Duration) {
	for {
		_, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, "status")
		if err != nil {
			time.Sleep(ZreplRetryInterval)
			klog.Errorf("zpool status returned error in zrepl startup : %v", err)
			klog.Infof("Waiting for zpool replication container to start...")
			continue
		}
		break
	}
}

// CheckForZreplContinuous is continuous health checker for status of zrepl in cstor-pool container.
func CheckForZreplContinuous(ZreplRetryInterval time.Duration) {
	for {
		out, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, "status")
		if err == nil {
			//even though we imported pool, it disappeared (may be due to zrepl container crashing).
			// so we need to reimport.
			if PoolAddEventHandled && strings.Contains(string(out), StatusNoPoolsAvailable) {
				break
			}
			time.Sleep(ZreplRetryInterval)
			continue
		}
		klog.Errorf("zpool status returned error in zrepl healthcheck : %v, out: %s", err, out)
		break
	}
}

// LabelClear is to clear zpool label on block devices.
func LabelClear(blockDevices []string) error {
	var failLabelClear = false
	for _, bd := range blockDevices {
		labelClearStr := []string{"labelclear", bd}
		stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, labelClearStr...)
		if err != nil {
			klog.Errorf("Unable to clear label on blockdevice %v: %v, err = %v", bd,
				string(stdoutStderr), err)
			failLabelClear = true
		} else {
			klog.Infof("successfully cleared label on blockdevice %v", bd)
		}
	}
	if failLabelClear {
		return fmt.Errorf("Unable to clear labels from all the blockdevices of the pool")
	}
	return nil
}

// GetDeviceIDs returns the list of device IDs for the csp.
func GetDeviceIDs(csp *apis.CStorPool) ([]string, error) {
	var bdDeviceID []string
	for _, group := range csp.Spec.Group {
		for _, blockDevice := range group.Item {
			bdDeviceID = append(bdDeviceID, blockDevice.DeviceID)
		}
	}
	if len(bdDeviceID) == 0 {
		return nil, errors.Errorf("No device IDs found on the csp %s", csp.Name)
	}
	return bdDeviceID, nil
}

// SetPoolRDMode set/unset pool readonly
func SetPoolRDMode(csp *apis.CStorPool, isROMode bool) error {
	mode := "off"
	if isROMode {
		mode = "on"
	}

	cmd := []string{"set",
		"io.openebs:readonly=" + mode,
		string(PoolPrefix) + string(csp.ObjectMeta.UID),
	}

	stdoutStderr, err := RunnerVar.RunCombinedOutput(zpool.PoolOperator, cmd...)
	if err != nil {
		return errors.Errorf("Failed to update readOnly mode out:%v err:%v", string(stdoutStderr), err)
	}
	return nil
}
