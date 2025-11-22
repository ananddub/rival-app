Project Setup & Website Plan
This document provides a systematic approach to building your website based on the provided A-Z Tech Stack. We will cover environment setup, folder structures, version control, and page layouts.
1. Core Web Tech Stack Summary
Framework: Next.js 14+ (with App Router)
Language: TypeScript
UI Library: TailwindCSS for styling, with shadcn/ui for pre-built, accessible components.
Data Tables: TanStack Table for powerful, headless tables.
Data Visualization: Recharts for interactive charts and graphs.
Forms: React Hook Form for performance, combined with Zod for robust validation.
Authentication: JSON Web Tokens (JWT) managed via secure HttpOnly cookies.
2. Initial Project & Environment Setup
This section covers the initial commands to create the project structure, which we have already completed.
Monorepo Root: E:\RIVL\RIVL_Website\
Applications:
E:\RIVL\RIVL_Website\apps\merchant-dashboard\
E:\RIVL\RIVL_Website\apps\admin-dashboard\
Shared Packages: E:\RIVL\RIVL_Website\packages\
3. GitHub Repository Setup
This section details how the code is saved and tracked. A root .gitignore file is used to exclude node_modules, .next, and other unnecessary files from version control. A README.md provides setup instructions for new developers.
4. Detailed Folder & File Structure
This is a more accurate blueprint for where all the code will live inside the merchant-dashboard app, reflecting our current progress.
apps/merchant-dashboard/
└── src/
    ├── app/
    │   ├── (auth)/             # Route group for authentication pages
    │   │   ├── login/
    │   │   │   └── page.tsx    # Login page component
    │   │   └── layout.tsx      # Layout specific to auth pages
    │   │
    │   ├── (dashboard)/        # Route group for main dashboard sections
    │   │   ├── dashboard/
    │   │   │   └── page.tsx    # Main dashboard view
    │   │   ├── orders/
    │   │   │   └── page.tsx    # Orders list page
    │   │   ├── customers/      # NEW: Customers section
    │   │   │   └── page.tsx    # Main customers list page
    │   │   ├── payouts/
    │   │   │   └── page.tsx    # Payouts history page
    │   │   ├── settings/
    │   │   │   └── page.tsx    # Settings page (likely tabbed)
    │   │   └── layout.tsx      # Shared layout for dashboard (header, sidebar)
    │   │
    │   ├── api/                # API routes
    │   │   └── auth/           # Example API route structure
    │   │       └── route.ts
    │   │
    │   ├── favicon.ico
    │   ├── globals.css         # Global styles
    │   └── layout.tsx          # Root layout for the entire application
    │
    ├── components/
    │   ├── ui/                 # Reusable low-level UI elements (from shadcn/ui)
    │   │   ├── button.tsx
    │   │   ├── card.tsx
    │   │   ├── sheet.tsx
    │   │   ├── input.tsx
    │   │   ├── label.tsx
    │   │   ├── table.tsx       # (Assuming shadcn table components used)
    │   │   ├── badge.tsx       # (Assuming shadcn badge component used)
    │   │   └── checkbox.tsx    # (Assuming shadcn checkbox component used)
    │   │   └── textarea.tsx    # (Assuming shadcn textarea component used)
    │   │
    │   ├── shared/             # Custom, shared components specific to this app
    │   │   ├── Sidebar.tsx
    │   │   ├── Header.tsx
    │   │   ├── Footer.tsx      # (If applicable)
    │   │   ├── CreateOfferModal.tsx # Specific shared component
    │   │   └── DateRangePicker.tsx # Specific shared component
    │   │
    │   └── icons.tsx           # Custom or wrapper icons if needed
    │
    └── lib/                    # Utility functions, hooks, etc.
        ├── utils.ts            # General utility functions (like cn)
        └── hooks.ts            # Custom React hooks (if needed)



5. Page Layouts & Component Plan
This section provides a detailed blueprint for each page within the RIVL web platform.
5.1. Merchant Dashboard (/dashboard)
Purpose: Provide an at-a-glance overview of business health, recent activity, and quick access to common actions.
Layout: A responsive multi-column/row grid that elegantly stacks on mobile devices, displayed within the main dashboard layout (header/sidebar). Background watermark. Glassmorphism cards.
Components:
Quick Actions Toolbar: Buttons for "Create New Offer" (opens modal) and "Export Report". Glowing Title.
Stat Cards Row: Summary cards for "Today's Revenue", "Today's Orders", "New Customers", "Loyalty Members".
Main Visuals Row: An area chart (neon glow) for "Revenue Overview" and a pie chart for "Payment Method Breakdown".
Secondary Insights Row: A bar chart (white glowing bars) for "Busiest Hours" and a vertical stack for "Top Customers" and "Live Activity Feed".
CreateOfferModal: Floating modal for creating new offers.
5.2. Merchant Orders Page (/orders)
Purpose: Allow merchants to view, search, and manage all customer orders.
Layout: A full-width page focused on a comprehensive data table, within the main dashboard layout.
Components:
Page Header: Title, Export button.
KPI Cards: "Today's Orders", "Pending", "Completed", "Average Order Value".
Filter/Search Bar: Search input, Status filter buttons, Date range picker button (opens modal).
Bulk Actions Toolbar (Conditional): Appears when rows are selected. Includes "Mark Completed", "Print Selected".
Data Table (Client-side implementation): Columns for Checkbox, Order ID, Customer, Date, Amount, Status (using Badge components). Sortable headers for Date, Amount, Status. Quick status update icons for Pending orders. Includes row actions (...) button.
Pagination Controls: Below the table.
Order Details View (Modal): Triggered by row action. Shows customer info, itemized list, total, order notes (editable), and context-aware action buttons (Cancel Order, Issue Refund, Contact Customer, Print).
Cancel Order Modal: Triggered from details modal. Includes reason dropdown.
Issue Refund Modal: Triggered from details modal. Includes amount input (full/partial) and reason dropdown.
Date Range Picker Modal: Contains the DateRangePicker component.
5.3. Merchant Customers Page (/customers)
Purpose: To serve as the merchant's central hub for viewing, searching, and managing their customer list. It provides insights into customer value, loyalty, and history, enabling merchants to build relationships and tailor offers effectively, aligning with the RIVL vision of merchant empowerment.
Layout: A full-width page displayed within the main dashboard layout (inheriting the header, sidebar, and background). It features a row of key metric cards at the top, followed by the main customer data table. Modals are used for viewing details and adding new customers.
Components:
Page Header: Contains the "Customers" title, a functional Search Bar (filters by name, email, phone), and a functional "Add Customer" Button which opens the AddCustomerModal.
KPI Cards: Three StatCard components showing "Total Customers," "New Customers (Month)," and "Loyalty Members."
Customer Table:
Displays the customer list within a Card.
Uses a table structure with sortable headers (Button components with icons) for Name, Join Date, Total Spent, and Loyalty Status.
Shows customer details per row (Name, Contact Info, Join Date, Total Spent).
Uses a LoyaltyBadge component to visually display the loyalty status in the table.
Includes a row action button (...) which opens the Customer Details Modal.
Implements Pagination controls (Button components and text) below the table.
Customer Details Modal: A modal triggered by the row action, displaying:
Sticky Header with customer name, email, and close button.
Profile Info section (Phone, Join Date, Total Spent, Loyalty Badge).
Order History section (listing order IDs).
Offers Redeemed section (listing offer names).
Customer Notes section (Textarea with a "Save Notes" button).
Sticky Footer with a functional "Send Offer" button (opens CreateOfferModal targeted at this customer).
Add Customer Modal (AddCustomerModal): A separate modal component with a form (Name, Email, Phone) to add new customers manually.
Create Offer Modal (CreateOfferModal): Reused from the dashboard, now accepts an optional targetCustomerId prop passed when triggered from the Customer Details Modal.
5.4. Merchant Payouts Page (/payouts)
Purpose: To provide merchants with a clear and transparent view of the earnings collected by RIVL on their behalf and transferred (paid out) to their bank account. It helps them track cash flow, understand fees, and reconcile their accounts.
Layout: A full-width page displayed within the main dashboard layout (inheriting the header, sidebar, and background). It features rows of summary cards at the top, followed by a dedicated card for the current payout cycle, and finally, the main data table for historical payouts.
Components:
Page Header: Contains the "Payouts" title and a functional "Date Range" button that opens a calendar modal.
Summary Cards: Three PayoutStatCard components displaying "Next Payout Amount" (estimated), "Last Payout Date" (with amount), and "Total Paid Out (YTD)".
Current Payout Cycle Card: A dedicated Card showing estimated Gross Sales, RIVL Fees, and Net Payout for the ongoing cycle, along with relevant dates.
Payout History Table:
Displays historical payouts within a Card.
Uses a table structure with sortable headers (Button components with icons) for Batch ID, Date Range, Gross Sales, RIVL Fees, Net Payout, and Status.
Shows payout details per row, including amounts color-coded for clarity (green for gross, red for fees).
Uses a PayoutStatusBadge component to visually display the status (e.g., "Paid").
Includes a functional "Download Statement" button (Button component) in each row that triggers a client-side CSV download.
Implements Pagination controls (Button components and text) below the table.
Date Range Picker Modal: A modal (Card containing DateRangePicker component) triggered by the header button, allowing filtering of the history table.
5.5. Merchant Settings Page (/settings)
Purpose: A centralized hub for merchants to configure their profile and offers.
Layout: A tabbed interface within the main dashboard layout.
Components (Tabs):
"Restaurant Profile": Form to edit name, address, hours, cuisine type, logo/banner image uploads.
"Offers & Discounts": Data table listing existing offers (link to edit/view), button to navigate to /offers/new or open "Create Offer" modal.
"Account Settings": Form to update merchant login email/password.
"Bank Details": Secure form/interface to manage bank account information for payouts.
5.6. Admin Dashboard Pages
This section outlines the plan for the RIVL team's internal management tool.
5.6.1. Admin Dashboard (/dashboard)
Purpose: Provide a high-level, real-time overview of the entire RIVL platform's health and key metrics.
Layout: A data-dense grid of stat cards, charts, and activity feeds.
Components:
Platform-Wide Stat Cards: "Total Active Merchants", "Total Users", "Total Transaction Volume (24h)", "Pending Merchant Approvals".
Key Charts: Line chart for "Platform Revenue Over Time" and a map visualization showing merchant distribution.
Quick Access Links: Buttons to jump to "Manage Merchants", "View Audit Logs", etc.
System Status Feed: A log showing important system events or potential issues.
5.6.2. Merchant Management (/merchants)
Purpose: Allow the RIVL team to onboard, view, and manage all merchants on the platform.
Layout: A full-width page centered around a powerful data table.
Components:
Page Header: Title, "Add New Merchant" button, and a search/filter bar.
Data Table: A list of all merchants with columns for Name, Status (Approved, Pending, Suspended), Onboarding Date, and Total Volume.
Row Actions: "View Details", "Approve", "Suspend", and "Edit".
5.6.3. User Management (/users)
Purpose: Provide a way to manage all end-user (customer) accounts.
Layout: A simple data table with search functionality.
Components:
Data Table: A list of all users with columns for User ID, Email/Phone, Sign-Up Date, and Status.
Row Actions: "View Wallet History" and "Suspend Account".
5.6.4. Transactions Overview (/transactions)
Purpose: A master log for viewing and searching all financial transactions across the platform for support and reconciliation.
Layout: A highly detailed data table with advanced filtering.
Components:
Advanced Filters: Filter by Merchant, User, Date Range, Transaction Type, and Status.
Data Table: A comprehensive log with columns for Transaction ID, Timestamp, Merchant, User, Amount, and RIVL Fee.
5.6.5. CMS (/cms)
Purpose: Allow non-technical team members to manage the content of static pages like "FAQ" and "Terms of Service".
Layout: A simple list-and-edit interface.
Components:
Page List: A table listing all manageable content pages (e.g., /faq, /terms).
Content Editor: A rich text editor (WYSIWYG) that opens when a page is selected for editing.
5.6.6. Audit Logs (/audit-logs)
Purpose: A critical security and compliance feature providing a read-only, tamper-evident log of all significant actions taken within the platform.
Layout: A simple, searchable log view.
Components:
Data Table: A log showing Timestamp, Actor (e.g., admin@rivl.com), Action (e.g., merchant.approve), and Target (Merchant ID: 123).
6. Development Plan (Current Phase)
✅ 6.1: Install UI & Utility Dependencies
✅ 6.2: Build Core Layout Components
✅ 6.3: Initialize shadcn/ui
✅ 6.4: Build Authentication Flow (Mocked)
➡️ 6.5: Build Dashboard Pages (In Progress)
✅ Enhance the main /dashboard page.
⬜ Build out the /orders page with a data table.
⬜ Build out the /payouts page.
⬜ Build out the /settings page with a tabbed interface.
7. Data Architecture
Source of Truth: The Core Data Model outlined in the Tech stack.pdf defines the structure of the PostgreSQL database.
Backend Service: A dedicated backend built in Go (Golang) will contain all business logic and be the sole interface to the database.
Frontend Interaction: This Merchant Dashboard website is the user interface for the data. It will make API calls to the Go backend to fetch data (e.g., orders for the Orders page) and to send data (e.g., updating restaurant info from the Settings page).
