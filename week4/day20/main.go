// Day 20: encoding/json — Marshal, Unmarshal, Struct Tags, Streaming
// HOW TO RUN: go run week4/day20/main.go
//
// Java dev key shifts:
//   - json:"fieldname" struct tag = @JsonProperty("fieldName") in Jackson
//   - json.Marshal = ObjectMapper.writeValueAsString() — Go []byte output
//   - json.Unmarshal = ObjectMapper.readValue() — requires pointer to target
//   - json.NewEncoder/NewDecoder — streaming (like Jackson's streaming API)
//   - json:"omitempty" — skip field if zero value (like @JsonInclude(NON_NULL))
//   - json:"-" — exclude field always (like @JsonIgnore)
//   - Custom marshal: implement json.Marshaler / json.Unmarshaler interfaces

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// === STRUCT TAGS ===
// Struct tag format: `json:"name,options"`
// Options: omitempty (skip if zero), - (always skip), string (number as string)

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Tags        []string  `json:"tags,omitempty"` // omit if nil/empty
	Internal    string    `json:"-"`               // always excluded from JSON
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Order struct {
	OrderID  string    `json:"order_id"`
	UserID   int       `json:"user_id"`
	Products []Product `json:"products"`
	Total    float64   `json:"total"`
	Status   string    `json:"status"`
}

// === CUSTOM MARSHAL / UNMARSHAL ===
// Java: @JsonSerialize / @JsonDeserialize with custom serializer

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String()) // "1h30m" instead of nanoseconds
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = dur
	return nil
}

type Config struct {
	Timeout Duration `json:"timeout"`
	MaxConn int      `json:"max_connections"`
}

// === INTERFACE FOR UNKNOWN STRUCTURES ===
// Java: Map<String, Object> for arbitrary JSON
// Go: map[string]interface{} or map[string]any

func main() {
	// === MARSHAL (struct → JSON) ===
	// Java: mapper.writeValueAsString(product)
	p := Product{
		ID:       1,
		Name:     "Go Programming Book",
		Price:    49.99,
		Tags:     []string{"golang", "programming"},
		Internal: "this won't appear in JSON",
		CreatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	// json.Marshal returns []byte
	data, err := json.Marshal(p)
	if err != nil {
		fmt.Println("marshal error:", err)
		return
	}
	fmt.Println("marshaled:", string(data))

	// Indented (pretty-print)
	pretty, _ := json.MarshalIndent(p, "", "  ")
	fmt.Println("\npretty:\n", string(pretty))

	// === UNMARSHAL (JSON → struct) ===
	// Java: mapper.readValue(json, Product.class)
	// MUST pass a pointer to the target
	jsonStr := `{"id":2,"name":"Effective Go","price":29.99,"tags":["golang"],"created_at":"2024-02-01T00:00:00Z"}`
	var p2 Product
	if err := json.Unmarshal([]byte(jsonStr), &p2); err != nil {
		fmt.Println("unmarshal error:", err)
		return
	}
	fmt.Printf("\nunmarshaled: %+v\n", p2)

	// === OMITEMPTY ===
	empty := Product{ID: 3, Name: "Minimal", Price: 9.99}
	data2, _ := json.Marshal(empty)
	fmt.Println("\nomitempty:", string(data2))
	// tags and description are omitted because they're zero values

	// === STREAMING with Encoder/Decoder ===
	// More efficient for large data or HTTP request/response bodies
	fmt.Println("\n--- streaming encode ---")
	var sb strings.Builder
	enc := json.NewEncoder(&sb)
	enc.SetIndent("", "  ")

	order := Order{
		OrderID:  "ORD-001",
		UserID:   42,
		Products: []Product{p, p2},
		Total:    79.98,
		Status:   "pending",
	}
	enc.Encode(order)
	fmt.Println(sb.String())

	// Decode from stream
	fmt.Println("--- streaming decode ---")
	orderJSON := `{"order_id":"ORD-002","user_id":99,"products":[],"total":0,"status":"new"}`
	dec := json.NewDecoder(strings.NewReader(orderJSON))
	var order2 Order
	dec.Decode(&order2)
	fmt.Printf("decoded order: %+v\n", order2)

	// === ARBITRARY JSON ===
	// When you don't know the structure (like Java's Map<String, Object>)
	arbitrary := `{"name":"Rahul","age":28,"skills":["go","java"],"active":true}`
	var m map[string]any
	json.Unmarshal([]byte(arbitrary), &m)
	fmt.Println("\narbitrary map:", m)
	fmt.Println("name:", m["name"])
	fmt.Println("age type:", fmt.Sprintf("%T", m["age"])) // float64! JSON numbers become float64
	skills := m["skills"].([]any) // type assert to use as slice
	fmt.Println("first skill:", skills[0])

	// === CUSTOM MARSHAL ===
	cfg := Config{
		Timeout: Duration{90 * time.Second},
		MaxConn: 100,
	}
	cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println("\ncustom marshal:\n", string(cfgJSON))

	var cfg2 Config
	json.Unmarshal(cfgJSON, &cfg2)
	fmt.Println("custom unmarshal timeout:", cfg2.Timeout.Duration)

	// === JSON ARRAY ===
	products := []Product{
		{ID: 1, Name: "A", Price: 10},
		{ID: 2, Name: "B", Price: 20},
	}
	arr, _ := json.Marshal(products)
	fmt.Println("\narray:", string(arr))

	var products2 []Product
	json.Unmarshal(arr, &products2)
	fmt.Println("decoded array len:", len(products2))

	// === NULL HANDLING ===
	// Use pointer fields for nullable JSON values
	type NullableUser struct {
		Name  string  `json:"name"`
		Email *string `json:"email"` // pointer = nullable
	}

	nu := NullableUser{Name: "Alice", Email: nil}
	nuJSON, _ := json.Marshal(nu)
	fmt.Println("\nnull field:", string(nuJSON)) // email: null

	email := "alice@example.com"
	nu2 := NullableUser{Name: "Bob", Email: &email}
	nu2JSON, _ := json.Marshal(nu2)
	fmt.Println("non-null field:", string(nu2JSON))
}

// === EXERCISES ===
// 1. Create a Config struct that reads from JSON:
//    {"database":{"host":"localhost","port":5432},"redis":{"host":"localhost","port":6379}}
//    Use nested structs with proper json tags.
//
// 2. Write a custom JSON marshaler for a Money type:
//    type Money struct { Amount int64; Currency string }
//    Marshal as "9.99 USD" (formatted string), unmarshal from same format.
//
// 3. Handle JSON with unknown keys gracefully:
//    Given a struct, use json.Decoder.DisallowUnknownFields() to reject
//    JSON with extra fields. When would you want this behavior?
//
// 4. Write a function jsonDiff(a, b []byte) ([]string, error) that:
//    - Decodes both JSON objects into map[string]any
//    - Returns a list of keys where values differ
//
// 5. What happens when you unmarshal a JSON number into an interface{}?
//    How do you avoid the float64 issue when you expect an integer?
//    Hint: json.Decoder.UseNumber()
