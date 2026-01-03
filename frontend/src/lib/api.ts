import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { StockCheckerService } from "../gen/stockchecker/v1/service_pb";

// Custom fetch that includes credentials for cross-origin cookie support
const fetchWithCredentials: typeof fetch = (input, init) => {
  return fetch(input, { ...init, credentials: "include" });
};

// Create a transport for the Connect client
const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_URL || "http://localhost:8080",
  fetch: fetchWithCredentials,
});

// Create the client
export const stockCheckerClient = createClient(StockCheckerService, transport);
