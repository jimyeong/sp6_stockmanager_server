## 🦉 Owlverload — Stock Management for Real Shops

| Item              | Details |
|--------------------|---------|
| 📅 Period          | Feb 2024 – Jan 2025 |
| 🧑‍💻 Role           | Sole developer (Planning → Backend → Frontend) |
| 🧰 Tech Stack      | React Native · Go (mux) · Redis · RDBMS |
| 🏪 Environment     | Real retail shop with unstable network coverage |

### 📝 Overview
Owlverload was born out of a **real operational problem** I encountered while supporting a local retail shop.  
The shop handled over 2,000 product types, each arriving in boxes with **different expiry dates**, making it increasingly difficult to track items approaching their sell-by dates.

To address this, I built a mobile application that allows staff to **register short-dated items** and view them in a **date-sorted list**, so they can check and act on them at the start of each shift.  
The system uses React Native for the mobile UI, Go (mux) for the backend, and an RDBMS for structured data storage.

### ⚡ Technical Challenges & Solutions
During actual use, we discovered **network dead zones** in the store. When staff used the app in these areas, **duplicate inserts and data inconsistencies** occurred frequently.  
Instead of relying on manual fixes, I engineered a **robust, systematic solution**:

- Designed **retry and offline-safe workflows** to gracefully handle unstable connectivity  
- Implemented a **Go + Redis–based idempotency layer** to guarantee **exactly-once insert semantics**, even under repeated requests  
- Ensured the **UI remained responsive and stateful**, even when the network temporarily dropped

### 🚀 Outcomes
- Resolved recurring duplicate data issues caused by unstable connectivity  
- Enabled staff to **quickly identify and manage short-dated items**, improving operational efficiency  
- Gained **hands-on experience designing architecture under real-world constraints**, focusing on robustness and reliability
