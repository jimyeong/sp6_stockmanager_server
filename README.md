# Owlverload – Expiry Tracking (Live Retail Project)

Owlverload is an expiry-date tracking system built and deployed in a live retail environment.
It helps staff identify stock items expiring within a configurable time window (7 / 14 / 21 / 30 days) so they can discount or rotate items before they expire.

> This repository contains legacy code from a real deployment.
> To demonstrate maintainable engineering practices, I extracted and refactored one production flow into a clean handler/service/repo structure (see **Clean Slice** below), while intentionally avoiding a full rewrite.

## What problem this solves

In many small/medium retail stores, expiry management is mostly reactive:
items are noticed after they expire, causing waste and operational risk.

Owlverload was built to make expiry work *visible and actionable* by surfacing “expiring soon” items directly from stock data.

## Evidence (live usage)

These screenshots are from a real store operation:

- Expiring Items screen showing items expiring within 30 days
- Discount shelf showing discounted products selected using the system

<table>
  <tr>
    <td width="50%">
      <b>1. Search Items</b><br/>
      Search products by code or name to retrieve item details.
      <br/><br/>
      <img src="https://github.com/user-attachments/assets/8eec94a9-8bb5-4ea1-a963-f326f5da9bd8" width="100%"/>
    </td>
    <td width="50%">
      <b>2. View Item Details</b><br/>
      Check barcode, multilingual names, current pricing status and stock condition.
      <br/><br/>
      <img src="https://github.com/user-attachments/assets/26b94669-75a9-4d39-8579-7d5829b4461c" width="100%"/>
    </td>
  </tr>
  <tr>
    <td width="50%">
      <b>3. Add Stock</b><br/>
      Update stock quantities by unit type (Box, Bundle, PCS).
      <br/><br/>
      <img src="https://github.com/user-attachments/assets/9c9f3ce2-3a93-48f5-b657-0abfd65f9d34" width="100%"/>
    </td>
    <td width="50%">
      <b>4. Check all expiring items with days left</b><br/>
      See them in one place altogether so you won't miss out chances to sell them.
      <br/><br/>
      <img src="https://github.com/user-attachments/assets/51a917ca-33b0-4045-a2c6-66a2b1916e7e" width="100%"/>
    </td>
  </tr>
  <tr>
	  <td colspan="2">
		  <b>5. Application</b><br/>
		  Take out the registered item and display them at On-Sale section.
		  <br/><br/>
		  <img src="https://github.com/user-attachments/assets/67584eac-3391-466f-b5d0-49152d60877d" width="100%"/>  
	  </td>	
  </tr>

</table>





## Key features

- List items expiring within N days (7/14/21/30)
- Stock-level tracking (stock ID, expiry date, discount info)
- Optional barcode / product image association (varies by data availability)
- Token validation middleware (Firebase)

## Clean Slice (refactored example)

The full codebase evolved in production and includes legacy structure.
To provide a high-signal example of maintainable code, the following flow is implemented using a clean separation of concerns:

- `GET /stocks/expiring?within=30`

Structure:
- `internal/http/handler`: request parsing + response formatting
- `internal/service`: expiry business logic (filtering, validation)
- `internal/repo`: DB access only
- `internal/domain`: core entities (no DB/HTTP dependencies)

> Why only one slice?
> A full rewrite would be costly and risky for a live system.
> I chose a minimal, high-signal refactor to demonstrate architecture and reasoning.

## Tech stack

- Go (HTTP API)
- SQL database (see `db/schema.sql`)
- Firebase token validation (see `firebase/`, `middleware/validate_token.go`)
- Cloudflare R2
- React Native

## Running locally

### 1) Configure environment
Create a `.env` file (or use `.env.staging`) and set:

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- Firebase credentials (if required)

### 2) Start API
```bash
go run main.go
