import { ethers } from "ethers";
import { CosmosClient } from "@cosmos-client/core";
import EventEmitter from "events";
import Queue from "bull";

// Configuration
const ETH_RPC_URL = process.env.ETH_RPC_URL || "https://sepolia.infura.io/v3/YOUR_INFURA_PROJECT_ID";
const COSMOS_RPC_URL = process.env.COSMOS_RPC_URL || "http://localhost:1317"; // Cosmos REST endpoint
const ETH_DUTCH_AUCTION_ADDRESS = process.env.ETH_DUTCH_AUCTION_ADDRESS || "0xYourDutchAuctionContractAddress";
const COSMOS_HTLC_MODULE = "htlc";

// Initialize Ethereum provider and contract interface with retry logic
const ethProvider = new ethers.providers.JsonRpcProvider(ETH_RPC_URL);
const dutchAuctionAbi = [
  "event NewOrder(uint256 orderId, address maker, uint256 amount, bytes32 hashLock)"
];
const dutchAuctionContract = new ethers.Contract(
  ETH_DUTCH_AUCTION_ADDRESS,
  dutchAuctionAbi,
  ethProvider
);

// Initialize Cosmos client
const cosmosClient = new CosmosClient({ baseUrl: COSMOS_RPC_URL });

// Event queue for orders and events
const orderQueue = new Queue("orderQueue");

// Event emitter for broadcasting to resolvers
const resolverEmitter = new EventEmitter();

// Listen for new order events on Ethereum with error handling and reconnection
function listenEthereumOrders(): void {
  dutchAuctionContract.on("NewOrder", (orderId: ethers.BigNumberish, maker: string, amount: ethers.BigNumberish, hashLock: string) => {
    console.log("New Ethereum order:", { orderId, maker, amount, hashLock });
    orderQueue.add({
      chain: "ethereum",
      orderId: orderId.toString(),
      maker,
      amount: amount.toString(),
      hashLock
    });
  });

  const provider = dutchAuctionContract.provider as ethers.providers.WebSocketProvider;

  provider._websocket.on("close", () => {
    console.error("Ethereum websocket closed. Attempting to reconnect...");
    setTimeout(() => {
      listenEthereumOrders();
    }, 5000);
  });

  provider._websocket.on("error", (err: Error) => {
    console.error("Ethereum websocket error:", err);
  });
}

// Listen for new order events on Cosmos HTLC module with retry and error handling
async function listenCosmosOrders(): Promise<void> {
  setInterval(async () => {
    try {
      const res = await cosmosClient.tx.searchAll({
        events: [`message.module='${COSMOS_HTLC_MODULE}'`],
        limit: 10
      });
      res.txs.forEach(tx => {
        tx.logs.forEach(log => {
          log.events.forEach(event => {
            if (event.type === "htlc_create") {
              const order = {
                chain: "cosmos",
                orderId: tx.txhash,
                maker: event.attributes.find(a => a.key === "sender")?.value,
                amount: event.attributes.find(a => a.key === "amount")?.value,
                hashLock: event.attributes.find(a => a.key === "hash_lock")?.value
              };
              console.log("New Cosmos order:", order);
              orderQueue.add(order);
            }
          });
        });
      });
    } catch (err) {
      console.error("Error fetching Cosmos HTLC events:", err);
    }
  }, 10000);
}

// Broadcast orders to resolvers
function broadcastOrders(): void {
  orderQueue.process(async (job) => {
    const order = job.data;
    console.log("Broadcasting order to resolvers:", order);
    resolverEmitter.emit("newOrder", order);
    // Implement actual broadcast logic here (e.g., WebSocket, HTTP)
  });
}

// Monitor escrow contract states and swap progress
const swapStatus = new Map<string, string>(); // orderId -> status (pending, deposited, claimed, expired)

function monitorEscrowStates(): void {
  console.log("Monitoring escrow contract states...");

  // Example: listen to claim/refund events on Ethereum and Cosmos and update swapStatus
  // This is a placeholder for actual event subscriptions and logic

  // Alert for failed or stuck swaps
  setInterval(() => {
    const now = Date.now();
    swapStatus.forEach((status, orderId) => {
      if (status === "pending") {
        console.warn(`Swap ${orderId} is still pending. Check for issues.`);
      }
      // Add more conditions as needed for stuck or expired swaps
    });
  }, 60000); // check every 60 seconds
}

// Main function to start relayer
async function startRelayer(): Promise<void> {
  listenEthereumOrders();
  listenCosmosOrders();
  broadcastOrders();
  monitorEscrowStates();
  console.log("Relayer service started.");
}

startRelayer().catch(console.error);

export {
  resolverEmitter,
  orderQueue,
  startRelayer
};
