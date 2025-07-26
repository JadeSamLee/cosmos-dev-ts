import keccak256 from "keccak256";
import { MerkleTree } from "merkletreejs";

interface Order {
  orderId: string;
  startTime: number;
  endTime: number;
  reservePrice: number;
  startPrice: number;
  totalAmount: number;
  merkleRoot?: string;
  hashLock?: string;
}

class MatchingEngine {
  private orders: Map<string, Order>;
  private filledAmounts: Map<string, number>;
  private secrets: Map<string, Set<string>>;

  constructor() {
    this.orders = new Map(); // orderId -> order details
    this.filledAmounts = new Map(); // orderId -> amount filled
    this.secrets = new Map(); // orderId -> Set of used secrets
  }

  // Add a new order
  addOrder(order: Order): void {
    this.orders.set(order.orderId, order);
    this.filledAmounts.set(order.orderId, 0);
    this.secrets.set(order.orderId, new Set());
  }

  // Calculate current price based on Dutch auction logic
  getCurrentPrice(orderId: string, currentTime: number): number {
    const order = this.orders.get(orderId);
    if (!order) throw new Error("Order not found");

    const elapsed = currentTime - order.startTime;
    const duration = order.endTime - order.startTime;
    if (elapsed >= duration) return order.reservePrice;

    const priceDiff = order.startPrice - order.reservePrice;
    const price = order.startPrice - (priceDiff * elapsed) / duration;
    return price;
  }

  // Attempt to fill a partial amount of an order
  fillOrder(orderId: string, amount: number, secret: Buffer): number {
    const order = this.orders.get(orderId);
    if (!order) throw new Error("Order not found");

    const filled = this.filledAmounts.get(orderId) ?? 0;
    if (filled + amount > order.totalAmount) {
      throw new Error("Fill amount exceeds order total");
    }

    // Verify secret with Merkle proof if applicable
    if (order.merkleRoot) {
      const leaf = keccak256(secret);
      const tree = new MerkleTree([], keccak256, { sortPairs: true });
      tree.root = Buffer.from(order.merkleRoot, "hex");

      // In real usage, the proof should be provided; here we assume verification
      // For demo, we skip proof verification or implement as needed
      // Check if secret already used
      const usedSecrets = this.secrets.get(orderId);
      if (!usedSecrets) {
        throw new Error("No secrets set for order");
      }
      if (usedSecrets.has(secret.toString("hex"))) {
        throw new Error("Secret already used");
      }
      usedSecrets.add(secret.toString("hex"));
    } else {
      // Single secret verification
      const hash = keccak256(secret);
      if (hash.toString("hex") !== order.hashLock) {
        throw new Error("Invalid secret");
      }
    }

    this.filledAmounts.set(orderId, filled + amount);
    // Return updated filled amount
    const updatedFilled = this.filledAmounts.get(orderId);
    if (updatedFilled === undefined) {
      throw new Error("Filled amount not found after update");
    }
    return updatedFilled;
  }

  // Prioritize orders based on profitability or other criteria
  prioritizeOrders(): Order[] {
    // Example: sort orders by highest remaining amount
    return Array.from(this.orders.values()).sort((a, b) => {
      const aRemaining = a.totalAmount - (this.filledAmounts.get(a.orderId) ?? 0);
      const bRemaining = b.totalAmount - (this.filledAmounts.get(b.orderId) ?? 0);
      return bRemaining - aRemaining;
    });
  }

  // Ensure atomicity and consistency (placeholder)
  atomicUpdate(orderId: string, updateFunc: () => void): void {
    // In real implementation, use locks or transactions
    updateFunc();
  }
}

export default MatchingEngine;
