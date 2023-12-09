import "./App.css";
import contracts from "./contracts/SecretSpend.json";
import { useEffect, useState } from "react";
import { Contract, ethers, toBigInt } from "ethers";

const contractAddress = "0x9AB81C32e1D621404b253c7fE0fC9972d1645E69";
const contractABI = contracts.abi;

function App() {
    const [currentAccount, setCurrentAccount] = useState(null);
    const [receiverID, setReceiverID] = useState("");
    const [amount, setAmount] = useState("");
    const [showToast, setShowToast] = useState("");

    const checkWalletIsConnected = () => {
        const { ethereum } = window;

        if (!ethereum) {
            console.log("Make sure you have MetaMask!");
            return;
        } else {
            console.log("We have the ethereum object", ethereum);
        }
    };

    const connectWalletHandler = async () => {
        const { ethereum } = window;

        if (!ethereum) {
            alert("Get MetaMask!");
            return;
        }

        try {
            const accounts = await ethereum.request({
                method: "eth_requestAccounts",
            });
            console.log("Connected", accounts[0]);
            setCurrentAccount(accounts[0]);
        } catch (error) {
            console.log(error);
        }
    };

    const proveAndTransferHandler = async (e) => {
        e.preventDefault();

        try {
            setShowToast("Generating Proof");
            setTimeout(() => setShowToast(""), 3000);

            const fromIndex = 0;
            const toIndex = receiverID;
            const transferAmount = amount;
            const url = `http://localhost:8080/transfer-funds?fromIndex=${fromIndex}&toIndex=${toIndex}&amount=${transferAmount}`;

            const response = await fetch(url, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            console.log("Response:", data);

            setShowToast("Proof Generated");
            setTimeout(() => setShowToast(""), 3000);

            if (window.ethereum && currentAccount) {
                const provider = new ethers.BrowserProvider(window.ethereum);
                const signer = await provider.getSigner();

                const contract = new Contract(
                    contractAddress,
                    contractABI,
                    signer
                );

                console.log("Contract object:", contract);

                const zkProof = {
                    proof: data.proof.map((x) => toBigInt(x)),
                    input: data.inputs.map((x) => toBigInt(x)),
                };

                const tx = await contract.transferPrivately(zkProof);
                await tx.wait();

                console.log("Transaction successful:", tx);

                // Show success toast
                setShowToast("Transfer Successful");
                setTimeout(() => setShowToast(""), 3000);
            } else {
                console.log(
                    "Ethereum object not found or no account connected"
                );
                // Handle error (e.g., show error toast)
            }
        } catch (error) {
            console.error("An error occurred during form submission:", error);
        }
    };

    const connectWalletButton = () => {
        return (
            <button
                onClick={connectWalletHandler}
                className="cta-button connect-wallet-button"
            >
                Connect Wallet
            </button>
        );
    };

    const transferForm = () => {
        const handleReceiverIDChange = (e) => {
            const value = Math.max(1, Math.min(31, Number(e.target.value)));
            setReceiverID(value);
        };

        const handleAmountChange = (e) => {
            setAmount(e.target.value);
        };

        const isFormValid = () => {
            return (
                receiverID !== "" &&
                amount !== "" &&
                !isNaN(receiverID) &&
                !isNaN(amount)
            );
        };

        return (
            <div className="form-container">
                {showToast && <div className="toast">{showToast}</div>}
                <form onSubmit={proveAndTransferHandler}>
                    <div className="form-row">
                        <label className="form-label">ReceiverID:</label>
                        <input
                            type="number"
                            value={receiverID}
                            onChange={handleReceiverIDChange}
                            className="form-input"
                            min="0"
                            max="31"
                        />
                    </div>

                    <div className="form-row">
                        <label className="form-label">Amount:</label>
                        <input
                            type="number"
                            value={amount}
                            onChange={handleAmountChange}
                            className="form-input"
                            min="1"
                        />
                    </div>
                    <button
                        type="submit"
                        className="submit-button"
                        disabled={!isFormValid()}
                    >
                        Prove & Transfer
                    </button>
                </form>
            </div>
        );
    };

    useEffect(() => {
        checkWalletIsConnected();
    }, []);

    return (
        <div className="main-app">
            <h1>$ecret $pend</h1>
            <div>{currentAccount ? transferForm() : connectWalletButton()}</div>
        </div>
    );
}

export default App;
