# YubiApp Frontend

A modern React frontend for the YubiApp authentication system, built with TypeScript, Tailwind CSS, React Query, and Vite.

## Features

- **YubiKey Authentication**: Secure device-based authentication using YubiKey OTP
- **Modern UI**: Beautiful, responsive interface with Tailwind CSS
- **Type Safety**: Full TypeScript support for better development experience
- **State Management**: React Query for efficient server state management
- **Dark Mode**: Built-in dark mode support
- **Real-time Updates**: Automatic data synchronization with the backend

## Tech Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Tailwind CSS** - Styling
- **React Query (TanStack Query)** - Server state management
- **Axios** - HTTP client
- **Vite** - Build tool and dev server
- **React Router DOM** - Client-side routing

## Prerequisites

- Node.js 18+ (recommended: Node.js 20+)
- npm or yarn
- YubiApp backend running on `http://localhost:8080`

## Installation

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
   ```bash
   npm run dev
   ```

4. Open your browser and navigate to `http://localhost:5173`

## Development

### Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint
- `npm run type-check` - Run TypeScript type checking

### Project Structure

```
frontend/
├── src/
│   ├── components/          # React components
│   │   └── YubiKeyAuth.tsx  # Main authentication component
│   │   └── YubiKeyAuth.tsx  # Main authentication component
│   ├── hooks/               # Custom React hooks
│   │   ├── useAuth.ts       # Authentication hooks
│   │   └── useDevices.ts    # Device management hooks
│   ├── services/            # API services
│   │   └── api.ts           # Main API service
│   ├── App.tsx              # Main app component
│   ├── main.tsx             # App entry point
│   └── index.css            # Global styles with Tailwind
├── public/                  # Static assets
├── index.html               # HTML template
├── package.json             # Dependencies and scripts
├── tailwind.config.js       # Tailwind configuration
├── postcss.config.js        # PostCSS configuration
├── tsconfig.json            # TypeScript configuration
└── vite.config.ts           # Vite configuration
```

### Key Components

#### YubiKeyAuth
The main authentication component that:
- Provides a clean interface for YubiKey OTP input
- Auto-submits when OTP reaches 44 characters
- Shows user information after successful authentication
- Handles logout functionality
- Displays error messages for failed authentication

#### API Service
Centralized API client that:
- Manages authentication tokens
- Provides type-safe API calls
- Handles error responses
- Automatically includes auth headers

#### React Query Hooks
Custom hooks for data fetching:
- `useAuthenticateDevice` - Handle device authentication
- `useCurrentUser` - Fetch current user data
- `useDevices` - Manage user devices
- `useRegisterDevice` - Register new devices
- `useDeregisterDevice` - Remove devices

## Configuration

### Backend URL
The frontend is configured to connect to the YubiApp backend at `http://localhost:8080/api/v1`. To change this:

1. Edit `src/services/api.ts`
2. Update the `baseURL` in the axios configuration

### Environment Variables
Create a `.env` file in the frontend directory for environment-specific configuration:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_TITLE=YubiApp
```

## Usage

### Authentication Flow

1. **Insert YubiKey**: Plug in your YubiKey device
2. **Generate OTP**: Tap the button on your YubiKey to generate a one-time password
3. **Auto-submit**: The OTP will be automatically submitted when it reaches 44 characters
4. **Success**: If authenticated, you'll see a greeting with your user information
5. **Logout**: Click "Sign Out" to log out

### Device Management

The frontend supports:
- Viewing registered devices
- Registering new devices
- Deregistering devices
- Device type support (YubiKey, TOTP, SMS, Email)

### Actions

Users can perform various actions:
- Sign in/out
- Notify HR about sick leave
- Notify HR about vacation
- Report travel status
- Update work location

## Styling

The project uses Tailwind CSS with custom components:

- `.btn-primary` - Primary button styling
- `.btn-secondary` - Secondary button styling
- `.input-field` - Form input styling
- `.card` - Card container styling

### Dark Mode
The application automatically supports dark mode based on system preferences. All components are designed to work seamlessly in both light and dark themes.

## API Integration

The frontend integrates with the YubiApp backend API:

- **Authentication**: `/auth/device` - Device-based authentication
- **Users**: `/users/*` - User management
- **Devices**: `/devices/*` - Device registration and management
- **Actions**: `/auth/action/*` - Performing authenticated actions
- **Resources**: `/resources/*` - Resource management
- **Roles**: `/roles/*` - Role management
- **Permissions**: `/permissions/*` - Permission management

## Error Handling

The application includes comprehensive error handling:

- Network errors are displayed to users
- Authentication failures show helpful messages
- Form validation prevents invalid submissions
- Loading states provide user feedback

## Security

- Authentication tokens are managed securely
- API calls include proper authorization headers
- Sensitive data is not stored in localStorage
- HTTPS is recommended for production

## Contributing

1. Follow the existing code style
2. Add TypeScript types for new features
3. Include error handling for new API calls
4. Test in both light and dark modes
5. Update documentation for new features

## Troubleshooting

### Common Issues

1. **Backend Connection**: Ensure the YubiApp backend is running on port 8080
2. **YubiKey Issues**: Make sure your YubiKey is properly inserted and working
3. **CORS Errors**: Check that the backend allows requests from the frontend origin
4. **Build Errors**: Ensure all dependencies are installed with `npm install`

### Development Tips

- Use the React Query DevTools for debugging API calls
- Check the browser console for detailed error messages
- Use the Network tab to inspect API requests
- Enable TypeScript strict mode for better type safety

## License

This project is part of the YubiApp system and follows the same license terms.

## User Activity CLI Commands

### List User Activity
```
yubiapp-cli user-activity list [flags]
```
**Flags:**
- `--from-datetime` (RFC3339)
- `--to-datetime` (RFC3339)
- `--user-ids` (comma-separated UUIDs)
- `--location-ids` (comma-separated UUIDs)
- `--status-ids` (comma-separated UUIDs)
- `--action-ids` (comma-separated UUIDs)
- `--limit` (default: 50)
- `--offset` (default: 0)

### User Activity Summary
```
yubiapp-cli user-activity summary --from-datetime <start> --to-datetime <end> [--user-ids <ids>]
```
Shows a summary report for the given period and users.

### User Activity for a Specific User
```
yubiapp-cli user-activity user <user-id> [flags]
```
**Flags:** (same as list)

**Example Output:**
```
Found 3 activities (showing 1-3 of 3):

ID: 123e4567-e89b-12d3-a456-426614174000
User: alice (alice@example.com)
Action: user-signin
From: 2023-01-01T09:00:00Z
To: 2023-01-01T17:00:00Z
Location: Main Office
Status: Signed In
Created: 2023-01-01T09:00:00Z
---
...
```
