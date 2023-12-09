import { HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
require("dotenv").config();

const config: HardhatUserConfig = {
    solidity: "0.8.19",
    networks: {
        scrollSepolia: {
            url: "https://sepolia-rpc.scroll.io/" || "",
            accounts:
                process.env.PRIVATE_KEY !== undefined
                    ? [process.env.PRIVATE_KEY]
                    : [],
        },
    },
    etherscan: {
        apiKey: {
            scrollSepolia: process.env.ETHERSCAN_API_KEY || "",
        },
        customChains: [
            {
                network: "scrollSepolia",
                chainId: 534351,
                urls: {
                    apiURL: "https://api-sepolia.scrollscan.com/api",
                    browserURL: "https://api.scrollscan.com/api",
                },
            },
        ],
    },
};

export default config;
