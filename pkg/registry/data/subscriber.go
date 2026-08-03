package data

import "sync"

type SubscriptionManager struct {
	subscribers map[string][]string
	mutex       sync.RWMutex
}

// NewSubscriptionManager 创建新的订阅管理器
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{subscribers: make(map[string][]string)}
}

// Subscribe 添加订阅
func (sm *SubscriptionManager) Subscribe(fileName, tag string, client string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	key := getKey(fileName, tag)
	if _, exists := sm.subscribers[key]; !exists {
		sm.subscribers[key] = []string{}
	}

	for _, existingClient := range sm.subscribers[key] {
		if existingClient == client {
			return
		}
	}

	sm.subscribers[key] = append(sm.subscribers[key], client)
}

// GetSubscribers 获取订阅文件的客户端列表
func (sm *SubscriptionManager) GetSubscribers(fileName, tag string) []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	key := getKey(fileName, tag)
	clients := sm.subscribers[key]
	result := make([]string, len(clients))
	copy(result, clients)
	return result
}

// Unsubscribe 取消订阅
func (sm *SubscriptionManager) Unsubscribe(fileName, tag string, client string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	key := getKey(fileName, tag)
	subscribers := sm.subscribers[key]
	for i, subscriber := range subscribers {
		if subscriber == client {
			sm.subscribers[key] = append(subscribers[:i], subscribers[i+1:]...)
			if len(sm.subscribers[key]) == 0 {
				delete(sm.subscribers, key)
			}
			break
		}
	}
}

// 判断文件是否订阅
func (sm *SubscriptionManager) IsSubscribed(fileName, tag string) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	key := getKey(fileName, tag)
	subscribers, exists := sm.subscribers[key]
	return exists && len(subscribers) > 0
}

// 获取订阅列表
func (sm *SubscriptionManager) GetSubscriptionList() [][]string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	list := make([][]string, 0, len(sm.subscribers))
	for key, subscribers := range sm.subscribers {
		entry := append([]string{key}, subscribers...)
		list = append(list, entry)
	}
	return list
}
