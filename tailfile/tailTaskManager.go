package tailfile

import (
	"github.com/sirupsen/logrus"
	"gocode/LogAgent_etcd/model"
)

var (
	collectEntryChan chan []model.CollectEntry  // 当etcd发现新配置时，通过该channel交互
	tailTaskMap      map[string]*model.TailTask // tailTaskMap用于存储path对应的tailTask
)

func InitTailTaskList(collectEntryList []model.CollectEntry) error {
	tailTaskMap = make(map[string]*model.TailTask, 16)
	tailTaskList, err := newTailTaskList(collectEntryList)
	if err != nil {
		logrus.Errorf("init tail failed, err = %v", err.Error())
		return err
	}
	// 初始化接受新配置的通信管道：collectEntryChan
	collectEntryChan = make(chan []model.CollectEntry, 8)
	logrus.Info("init tailFile success!")
	// 使用当前配置，读取日志文件
	useInitConfig(tailTaskList)
	// 启用新的协程来启用新配置
	go useNewConfig()
	return nil
}

// 创建tailTaskList
func newTailTaskList(collectEntryList []model.CollectEntry) ([]*model.TailTask, error) {
	// 打开文件
	var tailTaskList []*model.TailTask
	for _, collectEntry := range collectEntryList {
		tailFile, err := newTailFile(collectEntry.Path)
		if err != nil {
			logrus.Errorf("create tailFile failed, path = %v, err = %v", collectEntry.Path, err.Error())
			continue
		}
		tailTask := model.NewTailTask(tailFile, collectEntry.Topic)
		// 存入tailTaskMap
		tailTaskMap[collectEntry.Path] = tailTask
		tailTaskList = append(tailTaskList, tailTask)
	}
	return tailTaskList, nil
}

// PutCollectEntry 将新配置放入通信管道：collectEntryChan
func PutCollectEntry(collectEntryList []model.CollectEntry) {
	collectEntryChan <- collectEntryList
}

// 使用初始配置
func useInitConfig(tailTaskList []*model.TailTask) {
	for _, tailTask := range tailTaskList {
		go run(tailTask)
	}
}

// 使用新配置，新配置来源于etcd的watch
func useNewConfig() {
	for {
		useOneNewConfig()
	}
}

func useOneNewConfig() {
	select {
	case collectEntryList := <-collectEntryChan: // 拿到一次值，该函数就会执行完，并不是一直阻塞在当前步
		newConfigMap := map[string]struct{}{} // 新配置的map
		for _, collectEntry := range collectEntryList {
			// 创建新配置的map，方便第3步的查找
			newConfigMap[collectEntry.Path] = struct{}{}
			// 1. 配置已经存在，则维持
			if isExist(collectEntry.Path) {
				continue
			}
			// 2. 配置不存在，创建tail
			tailFile, err := newTailFile(collectEntry.Path)
			if err != nil {
				logrus.Errorf("create tailfile failed, err = %v", err.Error())
				continue
			}
			tailTask := model.NewTailTask(tailFile, collectEntry.Topic)
			tailTaskMap[collectEntry.Path] = tailTask // 放入map中进行维护
			go run(tailTask)
		}
		// 3. 老配置存在，新配置不存在
		deleteExistTailTask(newConfigMap)
	}
}

func isExist(path string) bool {
	_, exist := tailTaskMap[path]
	return exist
}

// 查找老配置中存在，但是新配置中不存在的配置，并在老配置map中删除
func deleteExistTailTask(newConfigMap map[string]struct{}) {
	logrus.Infof("findNotInTailTaskMap()执行了...")
	for key, _ := range tailTaskMap {
		if _, exist := newConfigMap[key]; !exist {
			// 老配置中不存在，删除对应tailTask
			logrus.Infof("the goroutine named : %s is ready to stop.", key)
			tailTask := tailTaskMap[key]
			delete(tailTaskMap, key)
			// 并且停掉对应的goroutine
			tailTask.Cancel() //结束tail.run()函数
		}
	}
}
