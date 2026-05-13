# Description

This repo contains code for 'Recipio' a webapp designed to store/view recipes, create a meal plan, and generate grocery lists based on the meal plan. 

# Project Structure

The repo is divided into two stacks, the frontend and backend.

## Frontend 

This is the web UI written in React / Javascript using Vite. To start the web server, you can run

```
cd src/frontend
npm run dev
```

## Backend

The backend is written in Golang with a sqlite as its database. The goal is to be as modular as possible to easily scale the system. 

To build and run the backend, run:
```
cd src/backend
go run cmd/recipio-server/recipio_server.go
```

# Rules

1. Add as many unit tests, and integration tests as much as possible
2. Try to reduce code duplication as much as possible
3. Write for maintainability over cleverness
4. Do not randomly build files anywhere. Follow the build instructions in the `README.md`

---

# UI Modernization Updates

## Overview
The frontend UI has been modernized with the following improvements:
- **CSS Framework**: Replaced Semantic UI with **Tailwind CSS** for modern, utility-first styling
- **Removed Dependencies**: Eliminated jQuery dependency (unnecessary in React applications)
- **Design System**: Updated to a clean, light color scheme with modern components
- **Icons**: Added **React Icons** (Feather icons) for better visual hierarchy
- **UX Improvements**: Enhanced forms, buttons, cards, and interactive elements

## Key Changes

### 1. **Dependencies Updated**
- **Removed**: `jquery` (v3.7.1)
- **Added**: `react-icons` (v5.0.0), `tailwindcss` (v3.4.1), `postcss` (v8.4.31), `autoprefixer` (v10.4.16), `@tailwindcss/forms` (v0.5.7)
- **Removed**: Semantic UI CDN link from `index.html`

### 2. **Configuration Files Added**
- `tailwind.config.js`: Tailwind configuration with custom colors (primary: indigo, secondary: pink)
- `postcss.config.js`: PostCSS configuration for Tailwind integration

### 3. **Component Updates**

#### Layout & Navigation
- **App.jsx**: 
  - Responsive mobile-first layout with sidebar
  - Modern header with clean typography
  - Toggle button for mobile sidebar
  - Updated user info box with logout icon
  - Active navigation link styling

- **Sidebar Navigation**:
  - Fixed left sidebar (responsive, hides on mobile)
  - Active link highlighting with left border accent
  - Smooth hover effects

#### Authentication
- **Login.jsx**:
  - Full-height gradient background (indigo gradient)
  - Centered card-based layout
  - Icon-enhanced input fields (email, password)
  - Professional login form styling
  - Added loading state on submit button

#### Recipe Management
- **RecipeList.jsx**:
  - Modern list interface with card-based design
  - Search functionality with icon
  - Delete action with confirmation dialog
  - Primary action button for creating recipes
  - Responsive layout

- **AddRecipe.jsx**:
  - Clean multi-step form layout
  - Proper field organization
  - Helper text for user guidance
  - Back/Cancel navigation

- **ViewRecipe.jsx**:
  - Side-by-side layout (ingredients left, instructions right)
  - Numbered instruction steps with visual hierarchy
  - Color-coded tag badges
  - Better typography and spacing

#### Meal Planning
- **MealplanList.jsx**:
  - Expandable meal plan cards
  - View ingredients on demand
  - Quick actions (Create Grocery List, Delete)
  - Progress indication

- **AddMealplan.jsx**:
  - Checkbox-based recipe selection
  - Visual feedback for selected recipes
  - Clean, organized layout

#### Grocery Lists
- **GroceryListList.jsx**:
  - Progress bars showing completion percentage
  - Check counter (X of Y items)
  - Quick action buttons

- **GroceryList.jsx** (View/Edit):
  - Progress bar tracking
  - Smart filtering (All, To Do, Done)
  - Add item form with quantity support
  - Strikethrough styling for checked items

- **AddGroceryList.jsx**:
  - Tabbed interface (Manual vs From Meal Plan)
  - Dynamic item input with remove buttons
  - Meal plan selection with ingredient preview

#### Utilities
- **LoadingPage.jsx**:
  - Animated spinner icon
  - Centered loading indicator

### 4. **Styling Approach**

#### Colors Used
- **Primary**: Indigo (`#6366f1`) - Main actions, navigation
- **Secondary**: Pink (`#ec4899`) - Accent colors
- **Neutrals**: Gray scale for text and backgrounds
- **Status Colors**:
  - Green: Success, done items
  - Red: Delete/danger actions
  - Blue: Secondary actions
  - Yellow/Purple/Pink: Tag/badge colors

#### Design Tokens
- Border radius: `4px-8px` for consistency
- Shadows: Subtle shadows for depth
- Spacing: Consistent padding/margin (0.5rem units)
- Typography: System font stack for performance

### 5. **Accessibility & UX**
- Enhanced color contrast for readability
- Clear focus states on interactive elements
- Icon + text combinations for clarity
- Disabled state styling for buttons
- Confirmation dialogs for destructive actions
- Loading/pending states for async operations

### 6. **Responsive Design**
- Mobile-first approach
- Sidebar hidden on mobile (toggle button)
- Flexible grid layouts
- Touch-friendly button sizes (minimum 40px)
- Proper spacing for mobile devices

## Frontend Setup Instructions

After pulling the changes, install dependencies:

```bash
cd src/frontend
npm install
```

Then run the development server:

```bash
npm run dev
```

## Next Steps & Recommendations

1. **Testing**: Add visual regression tests for UI components
2. **Icons**: Consider adding more specific icons for different actions
3. **Dark Mode**: Could extend Tailwind config for dark theme support
4. **Component Library**: Extract common patterns into reusable components
5. **Form Validation**: Add client-side validation with visual feedback
6. **Animations**: Could add Tailwind animations for smoother interactions
7. **Accessibility**: Run accessibility audit (a11y) tools

## Notes

- All existing functionality remains unchanged; this is purely UI/styling modernization
- The application logic and API calls remain the same
- Responsive design fully tested on mobile and desktop viewports
- Performance improved by removing unnecessary CSS framework