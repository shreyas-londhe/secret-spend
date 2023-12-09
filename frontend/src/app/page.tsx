"use client";
import { useState } from "react";
import styles from "./page.module.css";

export default function Home() {
    const [receiverId, setReceiverId] = useState("");
    const [amount, setAmount] = useState("");
    const [responseMessage, setResponseMessage] = useState(null);

    const handleSubmit = async (event: any) => {
        event.preventDefault(); // Prevents the default form submission behavior

        // Construct the API URL with query parameters
        const apiUrl = `http://localhost:8080/transfer-funds?fromIndex=0&toIndex=${encodeURIComponent(
            receiverId
        )}&amount=${encodeURIComponent(amount)}`;

        try {
            const response = await fetch(apiUrl, {
                method: "POST", // Assuming the API expects a POST request
            });

            if (response.ok) {
                // Handle the response if successful
                const data = await response.json();
                setResponseMessage(data);
                console.log("Funds transferred successfully:", data);
            } else {
                // Handle errors if the server response was not ok
                const errorData = await response.json();
                setResponseMessage(errorData);
                console.error("Server responded with an error:", errorData);
            }
        } catch (error) {
            // Handle errors if the fetch itself fails (e.g., network error)
            console.error("Failed to transfer funds:", error);
        }
    };

    return (
        <main className={styles.main}>
            <h1 className={styles.title}>SECRET SPEND</h1>
            <p className={styles.description}>Send funds to anyone privately</p>

            <form className={styles.form} onSubmit={handleSubmit}>
                <div className={styles.formRow}>
                    <label htmlFor="payto" className={styles.label}>
                        Receiver ID
                    </label>
                    <input
                        id="payto"
                        type="number"
                        placeholder="Enter Receiver ID"
                        className={styles.input}
                        value={receiverId}
                        onChange={(e) => setReceiverId(e.target.value)}
                        min="0"
                        max="31"
                    />
                </div>

                <div className={styles.formRow}>
                    <label htmlFor="amount" className={styles.label}>
                        Amount
                    </label>
                    <input
                        id="amount"
                        type="number"
                        placeholder="Enter Amount"
                        className={styles.input}
                        value={amount}
                        onChange={(e) => setAmount(e.target.value)}
                        min="0.01"
                        step="0.01"
                    />
                </div>

                <div className={styles.formRow}>
                    <button
                        type="submit"
                        className={styles.button}
                        disabled={!receiverId || !amount}
                    >
                        Prove & Transfer
                    </button>
                </div>
            </form>
        </main>
    );
}
