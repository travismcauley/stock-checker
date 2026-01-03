import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { StockCheckerService } from "../gen/stockchecker/v1/service_pb";

// Create a transport for the Connect client
const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_URL || "http://localhost:8080",
});

// Create the client
export const stockCheckerClient = createClient(StockCheckerService, transport);
