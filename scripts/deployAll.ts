import hre from "hardhat";

async function main(): Promise<void> {
  console.log("Starting deployment to Sepolia...");

  // Deploy DutchAuctionPartial
  const DutchAuctionPartial = await hre.ethers.getContractFactory("DutchAuctionPartial");
  const dutchAuction = await DutchAuctionPartial.deploy();
  await dutchAuction.deployed();
  console.log("DutchAuctionPartial deployed to:", dutchAuction.address);

  // Deploy EscrowFactory
  const EscrowFactory = await hre.ethers.getContractFactory("escrow/EscrowFactory");
  const escrowFactory = await EscrowFactory.deploy(
    "0xLimitOrderProtocolAddress", // Replace with actual address
    "0xFeeTokenAddress",           // Replace with actual fee token address
    "0xAccessTokenAddress",        // Replace with actual access token address
    "0xOwnerAddress",              // Replace with actual owner address
    3600,                         // rescueDelaySrc example
    3600                          // rescueDelayDst example
  );
  await escrowFactory.deployed();
  console.log("EscrowFactory deployed to:", escrowFactory.address);

  // Deploy Escrow
  const Escrow = await hre.ethers.getContractFactory("escrow/Escrow");
  const escrow = await Escrow.deploy();
  await escrow.deployed();
  console.log("Escrow deployed to:", escrow.address);

  // Verify contracts on Etherscan
  if (hre.network.name === "sepolia") {
    console.log("Verifying contracts on Etherscan...");
    try {
      await hre.run("verify:verify", {
        address: dutchAuction.address,
        constructorArguments: [],
      });
      await hre.run("verify:verify", {
        address: escrowFactory.address,
        constructorArguments: [
          "0xLimitOrderProtocolAddress",
          "0xFeeTokenAddress",
          "0xAccessTokenAddress",
          "0xOwnerAddress",
          3600,
          3600,
        ],
      });
      await hre.run("verify:verify", {
        address: escrow.address,
        constructorArguments: [],
      });
      console.log("Verification completed.");
    } catch (error) {
      console.error("Verification failed:", error);
    }
  }

  console.log("Deployment script completed.");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
