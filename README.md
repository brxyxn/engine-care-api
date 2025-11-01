# Engine Care API

## Getting started

### Prerequisites

- [Tilt](https://tilt.dev/) installed on your machine
- [Docker](https://www.docker.com/get-started/) installed and running
- [Kubernetes](https://kubernetes.io/docs/tasks/tools/) cluster (Context: Docker Desktop)

```bash
tilt up
```

This command will start the Tilt UI in your browser. From there, you can monitor the status of your services and view logs.

### Accessing the API

Once the services are up and running, you can access the Engine Care API using the Postman collection provided in 
the `docs` directory.:

1. Open Postman.
2. Import the `EngineCareAPI.postman_collection.json` file from the `docs` directory.
3. Enable the environment variables as needed, use local to get started.
4. Use the predefined requests to interact with the API.

### Stopping the services
To stop the services, simply run:

```bash
tilt down
```

This command will stop all the services that were started by Tilt.

---

## Project Structure

- cmd/: Main application entry points
- internal/: Core application logic and business rules
- pkg/: Shared libraries and utilities
- configs/: Configuration files
- docs/: Documentation and Postman collections

## API Workflow

1. Register/Login → get JWT
2. Create Org (admin) → get org_id
3. Add yourself as member → set X-Org-Id header
4. Create Customer → get customer_id
5. Add Vehicle → get vehicle_id
6. (Optional) Create Appointment → get appointment_id
7. Create Work Order → get work_order_id
8. (Optional) Link Appointment
9. Update WO Status → triggers event
10. Add/Update Items → triggers totals recalc
11. Send Notification
12. Complete Work Order

---

## API Endpoint Flow

### **Initial Setup Flow** (One-time per organization)

1. **POST `/auth/verify`** or **Stack Auth SDK call**
    - Creates Stack Auth user
    - Returns JWT token

2. **POST `/organizations`** (Service/Admin only)
    - Create organization
    - Returns `organization_id`
    - _Note: Typically done via admin panel or onboarding flow_

3. **POST `/organizations/:org_id/members`** (Owner/Admin)
    - Invite/add users to organization
    - Assign roles (`owner`, `admin`, `manager`, `mechanic`, `viewer`)
    - Returns membership record

### **Authentication Flow** (Every request)

4. **Headers for all subsequent requests:**
   ```
   Authorization: Bearer <stack_jwt_token>
   X-Org-Id: <organization_id>
   ```
    - Middleware validates token, upserts `app_users`, sets GUCs

### **Core Business Flow** (Typical work order lifecycle)

#### **Customer & Vehicle Management**

5. **POST `/customers`**
    - Create customer record
    - Returns `customer_id`

6. **POST `/customers/:customer_id/vehicles`**
    - Add vehicle for customer
    - Returns `vehicle_id`

7. **GET `/customers`** or **GET `/customers?search=...`**
    - List/search customers (for lookups)

#### **Appointment Scheduling** (Optional but recommended)

8. **POST `/appointments`**
    - Create appointment slot
    - Links `customer_id`, `vehicle_id`, `start_time`, `end_time`
    - Returns `appointment_id`

9. **PATCH `/appointments/:id`**
    - Update status: `pending` → `confirmed`

#### **Work Order Creation**

10. **POST `/work-orders`**
    - Create work order with items in one call
    - Include `customer_id`, `vehicle_id`, `items[]`
    - Triggers auto-calculate totals
    - Returns `work_order_id`

11. **POST `/appointments/:appointment_id/link`** (If appointment exists)
    - Body: `{ "work_order_id": "..." }`
    - Links appointment to work order

#### **Work Order Progression**

12. **PATCH `/work-orders/:id/status`**
    - Update status: `draft` → `new` → `scheduled` → `in_progress` → `completed`
    - Trigger auto-creates event in `work_order_events`

13. **POST `/work-orders/:id/items`**
    - Add labor/parts mid-job
    - Auto-recalculates totals

14. **PATCH `/work-orders/:id/items/:item_id`**
    - Update quantity, pricing
    - Auto-recalculates totals

15. **DELETE `/work-orders/:id/items/:item_id`**
    - Remove item
    - Auto-recalculates totals

#### **Events & Communication**

16. **GET `/work-orders/:id/events`**
    - View status history, notes, photos

17. **POST `/work-orders/:id/events`**
    - Add manual event (e.g., `note_added`, `photo_uploaded`)

18. **POST `/work-orders/:id/notify`**
    - Send notification to customer
    - Body: `{ "channel": "email|sms|whatsapp", "template_key": "status_update" }`
    - Logs to `notification_logs`

#### **Completion & Invoicing**

19. **PATCH `/work-orders/:id/status`**
    - Set to `ready_for_pickup` or `completed`
    - Sets `completed_at` timestamp

20. **GET `/work-orders/:id/invoice`** (Read-only calculated view)
    - Returns totals, line items for customer invoice

---

### **Additional Endpoints** (For full CRUD)

#### **Projects** (Optional grouping)
- **POST `/projects`** – Create project
- **GET `/projects`** – List projects
- **POST `/projects/:id/work-orders`** – Create WO under project

#### **Admin/Reporting**
- **GET `/work-orders?status=...&priority=...`** – Filter/search
- **GET `/dashboard/stats`** – Aggregate metrics
- **GET `/notifications`** – Notification history
