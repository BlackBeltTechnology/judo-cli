# Testing Guidelines

## Introduction

This comprehensive guide provides testing patterns and best practices for our TypeScript React application. It's designed to help developers and AI agents write effective, maintainable tests using Vitest and React Testing Library.

**Important Notes:**

- This is a **living document** that provides general testing patterns and approaches
- Code examples are **illustrative** and may not reflect the current state of the actual codebase
- Always refer to the actual implementation in your project when writing tests
- Adapt these patterns to fit your specific use cases and current code structure
- Some imports, type definitions, and API signatures may differ from your actual dependencies

---

## 1. Core Testing Principles

- Focus on testing behavior, not implementation details
- Use realistic mock data that matches actual API responses
- Test loading states, success states, and error conditions
- Verify user interactions and their side effects
- Test component rendering and re-rendering scenarios
- Verify hook state changes and return value updates
- Refrain from writing useless comments, e.g. technical ones explaining which block/method/function does what

## 2. Element Selection and Testing Best Practices

### Prefer data-testid for Reliable Testing

**Core Principle:** Use `data-testid` attributes for selecting interactive elements and dynamic content, and explicit text assertions for static content verification.

```typescript jsx
// ✅ Preferred: Using data-testid for reliable element selection
const submitButton = screen.getByTestId("submit-button");
const userForm = screen.getByTestId("user-form");
const loadingSpinner = screen.queryByTestId("loading-spinner");
const errorMessage = screen.queryByTestId("error-message");

// ✅ Good: Explicit text assertions for content verification
expect(screen.getByText("User successfully created")).toBeInTheDocument();
expect(screen.queryByText("Error occurred")).not.toBeInTheDocument();

// ❌ Avoid: Fragile selectors that depend on implementation details
const button = screen.getByRole("button", { name: /submit/i }); // Can break with text changes
const form = container.querySelector(".user-form"); // Can break with CSS changes
const input = screen.getByDisplayValue("John"); // Can break with data changes
```

### Component data-testid Guidelines

When creating or updating components, ensure they include appropriate data-testid attributes:

```typescript jsx
// Interactive elements should have data-testid
<Button data-testid="save-button" onClick={handleSave}>Save</Button>
<TextField data-testid="username-input" value={username} onChange={handleChange} />
<Select data-testid="role-select" value={role} onChange={handleRoleChange} />
<Dialog data-testid="confirmation-dialog" open={isOpen}>

// Key UI sections and dynamic content areas
<div data-testid="user-profile-section">
<LoadingSpinner data-testid="loading-spinner" />
<ErrorMessage data-testid="error-message">{error}</ErrorMessage>
<Table data-testid="users-table">

// Lists and repeating elements with dynamic identifiers
{users.map(user => (
  <div key={user.id} data-testid={`user-item-${user.id}`}>
    <Button data-testid={`edit-user-${user.id}`}>Edit</Button>
    <Button data-testid={`delete-user-${user.id}`}>Delete</Button>
  </div>
))}
```

### data-testid Naming Conventions

- Use kebab-case: `data-testid="user-profile-form"`
- Be descriptive: `data-testid="submit-user-form"` not `data-testid="btn"`
- Include context for dynamic elements: `data-testid="edit-user-123"`
- Group related elements: `data-testid="user-form-section"`, `data-testid="user-form-submit"`

### Testing Patterns with data-testid

```typescript jsx
// Testing form interactions
const usernameInput = screen.getByTestId("username-input");
const passwordInput = screen.getByTestId("password-input");
const submitButton = screen.getByTestId("login-submit");

await user.type(usernameInput, "john.doe");
await user.type(passwordInput, "password123");
await user.click(submitButton);
```

## 3. Required Test Structure and Mock Declarations

### **CRITICAL: Mock Declaration Requirements**

**ALL mocks MUST be declared using `vi.hoisted()` blocks.** This ensures proper hoisting and prevents timing issues with module mocking.

```typescript jsx
// ✅ REQUIRED: All mocks must be declared in vi.hoisted() blocks
const { mockHook, mockApi, mockComponent } = vi.hoisted(() => ({
  mockHook: vi.fn(),
  mockApi: vi.fn(),
  mockComponent: vi.fn(),
}));

// ❌ NEVER: Direct vi.fn() declarations at top level
const mockHook = vi.fn(); // This can cause timing issues and test failures
const mockApi = vi.fn(); // This pattern is not allowed
```

### Complete Test File Structure

```typescript jsx
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";

// ✅ MANDATORY: ALL mocks must be declared in vi.hoisted() blocks
const { mockHook, mockApi, mockToaster } = vi.hoisted(() => ({
  mockHook: vi.fn(),
  mockApi: vi.fn(),
  mockToaster: vi.fn(),
}));

// Module mocking with importOriginal pattern using hoisted mocks
vi.mock("~/hooks/useHook", async (importOriginal) => {
  const original = await importOriginal<typeof import("~/hooks/useHook")>();
  return { ...original, useHook: mockHook };
});

vi.mock("~/api/apiService", async (importOriginal) => {
  const original = await importOriginal<typeof import("~/api/apiService")>();
  return { ...original, apiCall: mockApi };
});

describe("ComponentName", () => {
  let queryClient: QueryClient;

  const createWrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions:
      {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    // Set happy path as default using hoisted mocks
    mockHook.mockReturnValue(defaultSuccessState);
    mockApi.mockResolvedValue({ data: mockData });
  });

  afterEach(() => {
    queryClient.clear();
  });

  it("should render component correctly", () => {
    // Test implementation
  });
});
```

### Vitest Configuration Notes
- **Only import what you need**: Import React Testing Library utilities, type definitions, and actual modules being tested
- **ALL mocks must use vi.hoisted()**: This is mandatory for proper test execution and prevents timing issues

### Mocking Parameter Types

When creating mock parameters for tests, always type them as `Partial<>` of the expected type and cast to the full type at usage:

```typescript
// Mock parameter with Partial typing
const mockFilterItem: Partial<GridFilterItem> = {
    id: "1",
    field: "status",
    operator: "equals",
    value: "active",
};

// Cast to full type when using
const result = someFunction(mockFilterItem as GridFilterItem);

// For arrays of mocked data
const mockFilterItems: Partial<GridFilterItem>[] = [
    { id: "1", field: "status", value: "active" },
    { id: "2", field: "name", value: "test" },
];

// Cast when using
const filterModel = { items: mockFilterItems as GridFilterItem[] };
```

## 4. Essential Test Patterns

### Environment & Global Mocking

Mocking ENV vars should be done via `vi.stubEnv` while mocking `window` attributes/functions should be done via `vi.stubGlobal`.

```typescript jsx
beforeEach(() => {
  vi.stubEnv("VITE_APP_MODE", "user");
  vi.stubGlobal("location", { origin: "http://localhost:3000" });
});

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
});
```

### Component Mocking

```typescript jsx
// Component mock declared in hoisted block
const { mockComponentName } = vi.hoisted(() => ({
  mockComponentName: vi.fn<FC<ComponentProps<any>>>(({children, ...props}) => (
    <div data-testid="mock-component" data-props={JSON.stringify(props)}>
      {children}
    </div>
  )),
}));

vi.mock("./ComponentName", async (importOriginal) => {
    const actual = await importOriginal<typeof import("./ComponentName")>();
    return {
        ...actual,
        ComponentName: mockComponentName,
    };
});
```

### Hook Testing & Re-rendering

```typescript jsx
// Test hook state changes
it("should update hook state when data changes", async () => {
  const { result, rerender } = renderHook(() => useCustomHook(initialParams), {
    wrapper: createWrapper,
  });

  // Initial state
  expect(result.current.isLoading).toBe(true);

  await waitFor(() => {
    expect(result.current.isSuccess).toBe(true);
  });

  // Change parameters to trigger re-render
  rerender({ newParams: "updated" });

  await waitFor(() => {
    expect(result.current.data).toEqual(updatedData);
  });
});
```

### Component Re-rendering

```typescript jsx
// Test component updates with changing props
it("should re-render when props change", () => {
  const initialProps = { userId: "user-1", permissions: ["read"] };
  const { rerender } = render(<Component {...initialProps} />);

  expect(screen.queryByTestId("admin-panel")).not.toBeInTheDocument();

  // Update props to trigger re-render
  const updatedProps = { userId: "user-1", permissions: ["read", "admin"] };
  rerender(<Component {...updatedProps} />);

  expect(screen.getByTestId("admin-panel")).toBeInTheDocument();
});
```

### **CRITICAL: TypeScript Type Safety in Tests**

**Always verify actual interface structure before creating mock objects.** This prevents runtime errors and ensures type safety.

#### Type Verification Best Practices

```typescript jsx
// ✅ REQUIRED: Import actual types and verify their structure
import type { GridFilterModel, GridFilterItem } from '@mui/x-data-grid-pro';

// ✅ CRITICAL: Always apply actual interface structure
const correctFilterModel: Partial<GridFilterModel> = {
    items: [],
    logicOperator: 'and', // ✅ Correct property name
};

// ❌ NEVER: Assume property names without verification
const incorrectFilterModel = {
  items: [],
  linkOperator: 'and', // This property doesn't exist on GridFilterModel
};
```

#### Interface Verification Steps

1. **Always import the actual type**: `import type { InterfaceName } from 'library'`
2. **Apply proper TypeScript typing**: Use imported types to leverage TypeScript's compile-time checking
3. **Create typed helper functions**: Don't inline object creation in tests
4. **Verify with TypeScript compiler**: Let TypeScript catch interface mismatches at compile time
5. **Test with realistic data**: Use data patterns that match actual usage


## 5. Test Coverage Requirements

1. Loading states - When data is being fetched
2. Success states - When data loads successfully
3. Error states - When operations fail
4. Permission-based rendering - Based on user permissions
5. User interactions - Button clicks, form submissions
6. Prop propagation - Data passed to child components
7. Conditional rendering - UI changes based on state
8. Hook state changes - Return value updates and side effects
9. Component re-rendering - Behavior when props/state change
10. Dynamic permission changes - Real-time permission updates

## 6. Naming Conventions

- Test files: `FileName.test.{tsx?}`
- Test IDs: `data-testid="kebab-case-description"`
- Mock variables: `mockFunctionName` or `mockComponentName`
- Describe blocks: Use component/hook names and group by functionality

## 7. Mandatory Assertions

- Verify loading spinners appear during async operations
- Check error toasts for failed operations (when applicable)
- Confirm callback functions called with correct parameters
- Validate conditional UI elements based on permissions/state
- Test that props reach child components via content or data attributes
- Verify hook return values change when expected
- Ensure components re-render correctly with new props/state

## 8. Complete Test Example with Best Practices

_A comprehensive example demonstrating our testing standards and patterns_

```typescript jsx
// External libraries - no default React import, use named exports
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';

// Type imports following our guidelines
import type { FC, ReactNode } from 'react';
import type { IAccount, IDashboard } from '~/types/api';

// ✅ MANDATORY: All mocks declared in vi.hoisted() blocks at the top of the file
const { mockUseAccountAPI, mockUseToaster, mockGetDashboards } = vi.hoisted(() => ({
    mockUseAccountAPI: vi.fn(),
    mockUseToaster: vi.fn(),
    mockGetDashboards: vi.fn(),
}));

// Mock modules before the main describe block using importOriginal pattern
// This pattern preserves the original module while replacing specific exports
vi.mock("~/hooks/useAccountAPI", async (importOriginal) => {
    const original = await importOriginal<typeof import("~/hooks/useAccountAPI")>();
    return {
        ...original,
        useAccountAPI: mockUseAccountAPI,
    };
});

vi.mock("~/hooks/useToaster", async (importOriginal) => {
    const original = await importOriginal<typeof import("~/hooks/useToaster")>();
    return {
        ...original,
        useToaster: mockUseToaster,
    };
});

vi.mock("~/api/dashboardService", async (importOriginal) => {
    const original = await importOriginal<typeof import("~/api/dashboardService")>();
    return {
        ...original,
        getDashboards: mockGetDashboards,
    };
});

// Test wrapper component with proper typing
const TestWrapper: FC<{ children: ReactNode; queryClient: QueryClient }> = ({ children, queryClient }) => (
    <QueryClientProvider client={queryClient}>
        <MemoryRouter>{children}</MemoryRouter>
    </QueryClientProvider>
);

describe("UserDashboard", () => {
    let queryClient: QueryClient;

    // Define mock return values at describe level for reusability across tests
    // Use Partial typing for mock data and cast at usage
    const mockAccount: Partial<IAccount> = {
        id: "user-123",
        name: "John Doe",
        permissions: ["read", "write"],
    };

    // Mock objects with multiple methods - these use the hoisted mocks
    const mockToasterMethods = {
        success: vi.fn(),
        error: vi.fn(),
        warning: vi.fn(),
    };

    // Use realistic data structures that match your actual API responses with Partial typing
    const mockDashboards: Partial<IDashboard>[] = [
        {
            id: "dashboard-1",
            title: "Sales Dashboard",
            isPublic: true,
        },
        {
            id: "dashboard-2",
            title: "Analytics Dashboard",
            isPublic: false,
        },
    ];

    beforeEach(() => {
        // Clear all mocks but don't redefine them - preserves mock structure
        vi.clearAllMocks();

        // Fresh QueryClient per test prevents cross-test data pollution
        queryClient = new QueryClient({
            defaultOptions: {
                queries: { retry: false }, // Faster test execution
                mutations: { retry: false },
            },
        });

        // Set default mock implementations using hoisted mocks - use happy path as baseline
        // Cast Partial mock data to full type at usage
        mockUseAccountAPI.mockReturnValue({
            accountQuery: {
                data: mockAccount as IAccount,
                isLoading: false,
                isSuccess: true,
                isError: false,
            },
        });

        mockUseToaster.mockReturnValue(mockToasterMethods);
        mockGetDashboards.mockResolvedValue({ data: mockDashboards as IDashboard[] });
    });

    afterEach(() => {
        queryClient.clear(); // Remove cached queries to prevent test interference
    });

    describe("Component Rendering", () => {
        // Test basic rendering with successful data loading
        it("should render user dashboards with correct props", async () => {
            render(
                <TestWrapper queryClient={queryClient}>
                    <UserDashboard />
                </TestWrapper>
            );

            // Wait for async data to load before making assertions
            await waitFor(() => {
                expect(screen.getByTestId("user-name")).toHaveTextContent(mockAccount.name!);
            });

            // Verify all expected elements are present
            expect(screen.getByTestId("dashboard-1")).toBeInTheDocument();
            expect(screen.getByTestId("dashboard-2")).toBeInTheDocument();

            // Verify props passed to child components via text content
            expect(screen.getByText("Sales Dashboard")).toBeInTheDocument();
            expect(screen.getByText("Analytics Dashboard")).toBeInTheDocument();
        });

        // Test loading state before data arrives
        it("should show loading state while fetching data", () => {
            // Override default mock to simulate loading state
            mockUseAccountAPI.mockReturnValue({
                accountQuery: {
                    data: null,
                    isLoading: true,
                    isSuccess: false,
                    isError: false,
                },
            });

            render(
                <TestWrapper queryClient={queryClient}>
                    <UserDashboard />
                </TestWrapper>
            );

            // Verify loading UI is shown and content is hidden
            expect(screen.getByTestId("loading-spinner")).toBeInTheDocument();
            expect(screen.queryByTestId("dashboard-list")).not.toBeInTheDocument();
        });
    });

    describe("User Interactions", () => {
        // Test form interactions and callback execution with proper data flow
        it("should handle dashboard creation with proper callback execution", async () => {
            const user = userEvent.setup();
            const mockOnCreate = vi.fn();

            render(
                <TestWrapper queryClient={queryClient}>
                    <UserDashboard onCreateDashboard={mockOnCreate} />
                </TestWrapper>
            );

            // Wait for component to be fully rendered
            await waitFor(() => {
                expect(screen.getByTestId("user-name")).toBeInTheDocument();
            });

            // Simulate user clicking create button
            await user.click(screen.getByTestId("create-dashboard-button"));

            // Fill in form data
            const nameInput = screen.getByTestId("dashboard-name-input");
            await user.type(nameInput, "New Dashboard");

            // Submit form
            await user.click(screen.getByTestId("save-dashboard-button"));

            // Verify callback is called with expected parameters including context data
            await waitFor(() => {
                expect(mockOnCreate).toHaveBeenCalledWith({
                    name: "New Dashboard",
                    userId: mockAccount.id, // Verify context data is passed correctly
                    permissions: mockAccount.permissions,
                });
            });
        });
    });

    describe("Conditional Rendering", () => {
        // Test permission-based UI rendering
        it("should hide admin features for read-only users", async () => {
            // Create user with limited permissions - maintaining Partial typing
            const readOnlyAccount: Partial<IAccount> = {
                ...mockAccount,
                permissions: ["read"], // Only read permission
            };

            // Override account mock for this specific test
            mockUseAccountAPI.mockReturnValue({
                accountQuery: {
                    data: readOnlyAccount as IAccount,
                    isLoading: false,
                    isSuccess: true,
                    isError: false,
                },
            });

            render(
                <TestWrapper queryClient={queryClient}>
                    <UserDashboard />
                </TestWrapper>
            );

            await waitFor(() => {
                expect(screen.getByTestId("user-name")).toBeInTheDocument();
            });

            // Admin features should not be visible for read-only users
            expect(screen.queryByTestId("create-dashboard-button")).not.toBeInTheDocument();
            expect(screen.queryByTestId("admin-panel")).not.toBeInTheDocument();

            // Read-only features should still be visible
            expect(screen.getByTestId("dashboard-list")).toBeInTheDocument();
        });
    });

    describe("Mock modifications mid testing", () => {
        // Test sequential API calls with different responses
        it("should handle multiple dashboard updates in sequence", async () => {
            const user = userEvent.setup();

            // Use Partial typing for mock data arrays
            const initialDashboards: Partial<IDashboard>[] = [
                { id: "dashboard-1", title: "Initial Dashboard", isPublic: false },
            ];

            // First API call returns initial data - use mockResolvedValueOnce for single use
            mockGetDashboards.mockResolvedValueOnce({ data: initialDashboards as IDashboard[] });

            render(
                <TestWrapper queryClient={queryClient}>
                    <UserDashboard />
                </TestWrapper>
            );

            // Verify initial data loads
            await waitFor(() => {
                expect(screen.getByText("Initial Dashboard")).toBeInTheDocument();
            });

            // Setup second API call with updated data for refresh scenario
            const updatedDashboards: Partial<IDashboard>[] = [
                { id: "dashboard-1", title: "Updated Dashboard", isPublic: true },
                { id: "dashboard-2", title: "New Dashboard", isPublic: false },
            ];
            mockGetDashboards.mockResolvedValueOnce({ data: updatedDashboards as IDashboard[] });

            // Trigger refresh action
            await user.click(screen.getByTestId("refresh-button"));

            // Verify updated data appears after refresh
            await waitFor(() => {
                expect(screen.getByText("Updated Dashboard")).toBeInTheDocument();
                expect(screen.getByText("New Dashboard")).toBeInTheDocument();
            });

            // Setup third API call with error to test error handling
            mockGetDashboards.mockRejectedValueOnce(new Error("Server down"));

            // Trigger another refresh to test error scenario
            await user.click(screen.getByTestId("refresh-button"));

            // Verify error handling while preserving previous data
            await waitFor(() => {
                expect(mockToasterMethods.error).toHaveBeenCalledWith(
                    "Failed to load dashboards",
                    "Please try again later"
                );
                // Previous data should still be visible during error state
                expect(screen.getByText("Updated Dashboard")).toBeInTheDocument();
            });
        });

        // Test dynamic permission changes during component lifecycle
        it("should react to external permission changes", async () => {
            let currentPermissions = ["read"];

            // Create a dynamic mock that returns current permissions - useful for simulating real-time updates
            // Use Partial typing for dynamic mock data
            mockUseAccountAPI.mockImplementation(() => ({
                accountQuery: {
                    data: {
                        id: "user-123",
                        name: "John Doe",
                        permissions: currentPermissions, // References mutable variable
                    } as IAccount, // Cast at usage for dynamic mock
                    isLoading: false,
                    isSuccess: true,
                    isError: false,
                },
            }));

            const { rerender } = render(
                <TestWrapper queryClient={queryClient}>
                    <UserDashboard />
                </TestWrapper>
            );

            // Verify initial read-only state
            await waitFor(() => {
                expect(screen.getByTestId("user-name")).toBeInTheDocument();
            });
            expect(screen.queryByTestId("admin-controls")).not.toBeInTheDocument();

            // Simulate external permission upgrade (e.g., admin grants new permissions)
            currentPermissions = ["read", "write", "admin"];
            rerender(
                <TestWrapper queryClient={queryClient}>
                    <UserDashboard />
                </TestWrapper>
            );

            // Verify admin controls appear after permission upgrade
            await waitFor(() => {
                expect(screen.getByTestId("admin-controls")).toBeInTheDocument();
            });

            // Simulate permission downgrade (e.g., admin revokes permissions)
            currentPermissions = ["read"];
            rerender(
                <TestWrapper queryClient={queryClient}>
                    <UserDashboard />
                </TestWrapper>
            );

            // Verify admin controls disappear after permission downgrade
            await waitFor(() => {
                expect(screen.queryByTestId("admin-controls")).not.toBeInTheDocument();
            });
        });
    });
});
```

## Key Takeaways

1. **ALWAYS use `vi.hoisted()`**: Every mock function must be declared in a `vi.hoisted()` block
2. **Use `data-testid` for reliable element selection**: Avoid fragile selectors that depend on text or CSS
3. **Test behavior, not implementation**: Focus on what the user experiences
4. **Cover all states**: Loading, success, error, and edge cases
5. **Use realistic mock data with Partial typing**: Type mock data as `Partial<T>` and cast to full type at usage
6. **Test user interactions**: Verify callbacks are called with correct parameters
7. **Test conditional rendering**: Ensure UI changes based on state/permissions
8. **Use proper async patterns**: `waitFor` for async operations, `userEvent` for interactions
9. **Mock at the right level**: Use `importOriginal` pattern to preserve module structure
10. **Clear mocks between tests**: Prevent test interference with `vi.clearAllMocks()`
