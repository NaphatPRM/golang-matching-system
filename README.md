# Location-Based Food Delivery Matching System

A Go-based system that emulates a food delivery matching system with real-time driver tracking and intelligent dispatch.

## Features

### Core Features
- **Restaurant Management**: Register restaurants with location and menu
- **Customer Management**: Register customers and manage delivery locations
- **Order System**: Place and track orders through the delivery lifecycle
- **Intelligent Dispatch Engine**: Assigns drivers based on:
  - Distance to restaurant
  - Current driver load (active jobs)
  - Driver availability status

### Real-time Features
- Real-time driver location updates
- WebSocket support for live location tracking
- Automatic order dispatching

### Applications
- REST API Server
- Driver CLI Simulation App

## Project Structure

```
.
├── cmd/
│   ├── server/          # Main API server
│   └── driver-cli/      # Driver simulation CLI app
├── internal/
│   └── dispatch/        # Dispatch engine (driver assignment logic)
├── pkg/
│   ├── api/            # HTTP handlers
│   ├── models/         # Data models
│   └── services/       # Business logic services
├── go.mod
└── README.md
```

## Getting Started

### Prerequisites
- Go 1.21 or later

### Installation

```bash
# Clone the repository
git clone https://github.com/NaphatPRM/golang-matching-system.git
cd golang-matching-system

# Download dependencies
go mod download

# Build the applications
go build ./...
```

### Running the Server

```bash
# Start the API server (default port: 8080)
go run cmd/server/main.go

# Or specify a custom port
PORT=3000 go run cmd/server/main.go
```

### Running the Driver CLI App

```bash
# Start the driver CLI app (connects to localhost:8080 by default)
go run cmd/driver-cli/main.go

# Or specify a custom server URL
SERVER_URL=http://localhost:3000/api/v1 go run cmd/driver-cli/main.go
```

## API Endpoints

### Health Check
- `GET /health` - Server health check

### Restaurants
- `POST /api/v1/restaurants` - Register a new restaurant
- `GET /api/v1/restaurants` - Get all restaurants
- `GET /api/v1/restaurants/{id}` - Get a specific restaurant
- `PUT /api/v1/restaurants/{id}/menu` - Update restaurant menu
- `PUT /api/v1/restaurants/{id}/status` - Set restaurant open/closed status
- `GET /api/v1/restaurants/nearby?lat={lat}&lon={lon}&radius={km}` - Get nearby restaurants

### Customers
- `POST /api/v1/customers` - Register a new customer
- `GET /api/v1/customers` - Get all customers
- `GET /api/v1/customers/{id}` - Get a specific customer
- `PUT /api/v1/customers/{id}/location` - Update customer location

### Drivers
- `POST /api/v1/drivers` - Register a new driver
- `GET /api/v1/drivers` - Get all drivers
- `GET /api/v1/drivers/{id}` - Get a specific driver
- `PUT /api/v1/drivers/{id}/location` - Update driver location
- `PUT /api/v1/drivers/{id}/status` - Update driver status (offline/available/busy)
- `GET /api/v1/drivers/available` - Get all available drivers
- `GET /api/v1/drivers/locations/ws` - WebSocket endpoint for real-time location updates

### Orders
- `POST /api/v1/orders` - Place a new order
- `GET /api/v1/orders` - Get all orders
- `GET /api/v1/orders/{id}` - Get a specific order
- `PUT /api/v1/orders/{id}/status` - Update order status
- `POST /api/v1/orders/{id}/dispatch` - Manually dispatch an order
- `GET /api/v1/orders/pending` - Get all pending orders
- `GET /api/v1/orders/customer/{customerId}` - Get orders by customer
- `GET /api/v1/orders/driver/{driverId}` - Get orders by driver

### Dispatch
- `POST /api/v1/dispatch/scores` - Get driver scores for a pickup location

## Example Usage

### 1. Register a Restaurant
```bash
curl -X POST http://localhost:8080/api/v1/restaurants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Pizza Palace",
    "location": {"latitude": 13.7563, "longitude": 100.5018},
    "menu": [
      {"name": "Margherita Pizza", "price": 12.99},
      {"name": "Pepperoni Pizza", "price": 14.99}
    ]
  }'
```

### 2. Register a Customer
```bash
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "location": {"latitude": 13.7600, "longitude": 100.5100}
  }'
```

### 3. Register a Driver
```bash
curl -X POST http://localhost:8080/api/v1/drivers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Driver Mike",
    "max_concurrent": 3
  }'
```

### 4. Set Driver as Available
```bash
curl -X PUT http://localhost:8080/api/v1/drivers/{driver_id}/status \
  -H "Content-Type: application/json" \
  -d '{"status": "available"}'
```

### 5. Update Driver Location
```bash
curl -X PUT http://localhost:8080/api/v1/drivers/{driver_id}/location \
  -H "Content-Type: application/json" \
  -d '{"location": {"latitude": 13.7563, "longitude": 100.5018}}'
```

### 6. Place an Order
```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "{customer_id}",
    "restaurant_id": "{restaurant_id}",
    "items": [
      {"name": "Margherita Pizza", "price": 12.99, "quantity": 2}
    ],
    "delivery_location": {"latitude": 13.7600, "longitude": 100.5100}
  }'
```

The dispatch engine will automatically assign an available driver to the order.

## Driver CLI Features

The driver CLI app provides an interactive interface for drivers to:

1. **Register/Login**: Create a new driver account or login to existing one
2. **View Info**: See current status, location, and active orders
3. **Update Location**: Manually set location or use random coordinates
4. **Go Online/Offline**: Change availability status
5. **View Orders**: See assigned orders and pending orders
6. **Update Order Status**: Mark orders as picked up, delivering, delivered, or cancelled
7. **Simulate Movement**: Automatically send location updates to simulate driving

## Dispatch Algorithm

The dispatch engine uses a weighted scoring system:

```
Score = (DistanceWeight × NormalizedDistance) + (LoadWeight × NormalizedLoad)
```

Where:
- **NormalizedDistance**: 1 = closest, 0 = farthest (within search radius)
- **NormalizedLoad**: 1 = no active orders, 0 = at maximum capacity
- Default weights: Distance = 0.6, Load = 0.4

## Order Status Flow

```
pending → assigned → picked_up → delivering → delivered
    ↓          ↓           ↓            ↓
    └──────────┴───────────┴────────────┴──→ cancelled
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests with coverage
go test ./... -cover
```

## Configuration

### Dispatch Engine Configuration
- `MaxSearchRadiusKm`: Maximum radius to search for drivers (default: 10km)
- `DistanceWeight`: Weight for distance factor (default: 0.6)
- `LoadWeight`: Weight for load factor (default: 0.4)
- `AutoDispatchInterval`: Interval for auto-dispatching (default: 5s)
- `MaxDriversToConsider`: Max drivers to evaluate (default: 10)

## License

MIT License
