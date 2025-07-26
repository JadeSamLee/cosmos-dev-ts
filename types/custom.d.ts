declare module '@cosmos-client/core' {
  // Minimal declarations for @cosmos-client/core
  export class CosmosClient {
    constructor(options: { baseUrl: string });
    tx: {
      searchAll(params: any): Promise<any>;
    };
  }
  export class SigningStargateClient {
    static connectWithMnemonic(baseUrl: string, mnemonic: string): Promise<SigningStargateClient>;
    signAndBroadcast(senderAddress: string, msgs: any[], fee: any): Promise<any>;
  }
}

declare module 'keccak256' {
  function keccak256(data: string | Buffer): Buffer;
  export = keccak256;
}

declare module 'merkletreejs' {
  export class MerkleTree {
    constructor(leaves: Buffer[], hashFunction: (data: Buffer) => Buffer, options?: { sortPairs: boolean });
    root: Buffer;
    getHexRoot(): string;
    getProof(leaf: Buffer): Buffer[];
    verify(proof: Buffer[], leaf: Buffer, root: Buffer): boolean;
  }
}
