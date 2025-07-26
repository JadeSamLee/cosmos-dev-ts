import { expect } from "chai";
import { ethers } from "hardhat";
import { MerkleTree } from "merkletreejs";
import keccak256 from "keccak256";

describe("DutchAuctionPartial", function () {
  let DutchAuctionPartial: any, dutchAuction: any, token: any, owner: any, resolver1: any, resolver2: any;
  let secrets: string[], leaves: Buffer[], merkleTree: MerkleTree, merkleRoot: string;

  beforeEach(async function () {
    const signers = await ethers.getSigners();
    owner = signers[0];
    resolver1 = signers[1];
    resolver2 = signers[2];
    const other = signers[3];

    // Deploy a mock ERC20 token
    const Token = await ethers.getContractFactory("ERC20Mock");
    token = await Token.deploy("TestToken", "TTK", owner.address, ethers.utils.parseEther("1000"));
    await token.deployed();

    // Deploy DutchAuctionPartial contract
    DutchAuctionPartial = await ethers.getContractFactory("DutchAuctionPartial");
    dutchAuction = await DutchAuctionPartial.deploy();
    await dutchAuction.deployed();

    // Prepare secrets and Merkle tree
    secrets = ["secret1", "secret2", "secret3"].map(s => ethers.utils.keccak256(ethers.utils.toUtf8Bytes(s)));
    leaves = secrets.map(s => Buffer.from(s.slice(2), "hex"));
    merkleTree = new MerkleTree(leaves, keccak256, { sortPairs: true });
    merkleRoot = merkleTree.getHexRoot();

    // Approve tokens for dutchAuction contract
    await token.connect(owner).approve(dutchAuction.address, ethers.utils.parseEther("100"));
  });

  it("should create an order", async function () {
    const tx = await dutchAuction.createOrder(
      token.address,
      ethers.utils.parseEther("100"),
      ethers.utils.parseEther("10"),
      ethers.utils.parseEther("1"),
      3600,
      merkleRoot
    );
    await expect(tx).to.emit(dutchAuction, "OrderCreated");
  });

  it("should allow partial fill with valid secret and proof", async function () {
    await dutchAuction.createOrder(
      token.address,
      ethers.utils.parseEther("100"),
      ethers.utils.parseEther("10"),
      ethers.utils.parseEther("1"),
      3600,
      merkleRoot
    );

    const orderId = 0;
    const secret = secrets[0];
    const proof = merkleTree.getHexProof(secret);

    await expect(dutchAuction.connect(resolver1).partialFill(orderId, secret, proof))
      .to.emit(dutchAuction, "PartialFill")
      .withArgs(orderId, resolver1.address, await dutchAuction.getCurrentPrice(orderId), secret);
  });

  it("should reject partial fill with invalid secret or proof", async function () {
    await dutchAuction.createOrder(
      token.address,
      ethers.utils.parseEther("100"),
      ethers.utils.parseEther("10"),
      ethers.utils.parseEther("1"),
      3600,
      merkleRoot
    );

    const orderId = 0;
    const invalidSecret = ethers.utils.keccak256(ethers.utils.toUtf8Bytes("invalid"));
    const proof = merkleTree.getHexProof(invalidSecret);

    await expect(dutchAuction.connect(resolver1).partialFill(orderId, invalidSecret, proof)).to.be.revertedWith("Invalid Merkle proof");
  });

  it("should prevent reuse of secrets", async function () {
    await dutchAuction.createOrder(
      token.address,
      ethers.utils.parseEther("100"),
      ethers.utils.parseEther("10"),
      ethers.utils.parseEther("1"),
      3600,
      merkleRoot
    );

    const orderId = 0;
    const secret = secrets[0];
    const proof = merkleTree.getHexProof(secret);

    await dutchAuction.connect(resolver1).partialFill(orderId, secret, proof);
    await expect(dutchAuction.connect(resolver2).partialFill(orderId, secret, proof)).to.be.revertedWith("Secret already used");
  });

  it("should allow winner to claim order", async function () {
    await dutchAuction.createOrder(
      token.address,
      ethers.utils.parseEther("100"),
      ethers.utils.parseEther("10"),
      ethers.utils.parseEther("1"),
      3600,
      merkleRoot
    );

    const orderId = 0;
    const secret = secrets[0];
    const proof = merkleTree.getHexProof(secret);

    await dutchAuction.connect(resolver1).partialFill(orderId, secret, proof);

    await expect(dutchAuction.connect(resolver1).claimOrder(orderId))
      .to.emit(dutchAuction, "OrderClaimed")
      .withArgs(orderId, resolver1.address);
  });

  it("should reject claim by non-winner", async function () {
    await dutchAuction.createOrder(
      token.address,
      ethers.utils.parseEther("100"),
      ethers.utils.parseEther("10"),
      ethers.utils.parseEther("1"),
      3600,
      merkleRoot
    );

    const orderId = 0;
    const secret = secrets[0];
    const proof = merkleTree.getHexProof(secret);

    await dutchAuction.connect(resolver1).partialFill(orderId, secret, proof);

    await expect(dutchAuction.connect(resolver2).claimOrder(orderId)).to.be.revertedWith("Not winner");
  });

  // Example of mocking an external contract (ERC20)
  it("should mock external ERC20 contract", async function () {
    const MockERC20 = await ethers.getContractFactory("ERC20Mock");
    const mockToken = await MockERC20.deploy("MockToken", "MTK", owner.address, ethers.utils.parseEther("1000"));
    await mockToken.deployed();

    expect(await mockToken.balanceOf(owner.address)).to.equal(ethers.utils.parseEther("1000"));
  });
});
