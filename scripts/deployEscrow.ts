import hre from "hardhat";

async function main(): Promise<void> {
  // Replace with the actual deployed factory address on Sepolia
  const factoryAddress = "0xYourEscrowFactoryAddressHere";

  // Get the IEscrowFactory interface contract factory
  const IEscrowFactory = await hre.ethers.getContractFactory("IEscrowFactory");

  // Attach to the deployed factory contract
  const factory = IEscrowFactory.attach(factoryAddress);

  console.log("Deploying new escrow contract via factory...");

  // Call deployEscrow function on the factory to deploy a new escrow contract
  const tx = await factory.deployEscrow();
  const receipt = await tx.wait();

  // Assuming the factory emits an event with the new escrow address, parse it here
  // For demonstration, just log the transaction receipt
  console.log("Transaction receipt:", receipt);

  // TODO: Extract the new escrow contract address from the event logs if available
  // const newEscrowAddress = ...;

  // console.log("New escrow contract deployed at:", newEscrowAddress);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
