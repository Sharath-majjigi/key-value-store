package request

import(
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/gofiber/fiber/v2"
)

type KeyValue struct {
	value      []string
	expiryTime *time.Time
}

type Database struct {
	M   map[string]*KeyValue
	mux sync.RWMutex
}

func HandleSetCommand(c *fiber.Ctx, db *Database, args []string) error {
	if len(args) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid command",
		})
	}

	key := args[0]
	value := args[1]
	expiryTime := time.Time{}
	condition := ""

	for i := 2; i < len(args); i += 2 {
		switch strings.ToUpper(args[i]) {
		case "EX":
				if i+1 < len(args) {
					if seconds, err := strconv.Atoi(args[i+1]); err == nil {
						expiryTime = time.Now().Add(time.Duration(seconds) * time.Second)
					}
				}
		case "NX":
			condition = "NX"
		case "XX":
			condition = "XX"
		}
	}

	db.mux.Lock()
	defer db.mux.Unlock()

	if condition == "NX" && db.M[key] != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "key already exists",
		})
	}

	if condition == "XX" && db.M[key] == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "key not found",
		})
	}

	db.M[key] = &KeyValue{
		value:      []string{value},
		expiryTime: &expiryTime,
	}

	// println(expiryTime.Before(time.Now()))
	// println(expiryTime.After(time.Now()))
	// return c.JSON(fiber.Map{
	// 	"key":key,
	// 	"value":value,
	// 	"expiryTime":expiryTime,
	// })
	return nil
}

func HandleGetCommand(c *fiber.Ctx, db *Database, args []string) error {
	if len(args) != 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid command",
		})
	}

	key := args[0]

	db.mux.RLock()
	defer db.mux.RUnlock()

	if kv, ok := db.M[key]; ok {
		expt:=kv.expiryTime
		if expt != nil && time.Now().Before(*expt) {
			delete(db.M, key)

			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "key not found",
				})
		}
		return c.JSON(fiber.Map{
			"value": kv.value,
		})
	}
	
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "key not found",
	})
}

func HandleQPushCommand(c *fiber.Ctx, db *Database, args []string) error {
	if len(args) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid command",
		})
	}

	key := args[0]
	values := args[1:]

	db.mux.Lock()
	defer db.mux.Unlock()

	kv, ok := db.M[key]
	if !ok {
		kv = &KeyValue{}
		db.M[key] = kv
	}

	kv.value = append(kv.value, values...)

	return nil
}

func HandleQPopCommand(c *fiber.Ctx, db *Database, args []string) error {
	if len(args) != 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid command",
		})
	}

	key := args[0]

	db.mux.Lock()
	defer db.mux.Unlock()

	if kv, ok := db.M[key]; ok {
		if len(kv.value) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "queue is empty",
			})
		}

		value := kv.value[len(kv.value)-1]
		kv.value = kv.value[:len(kv.value)-1]

		return c.JSON(fiber.Map{
			"value": value,
		})
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "key not found",
	})
}