// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title OXGToken
 * @dev Token ERC-20 nativo de Oxy•gen Economy para la blockchain nativa
 * 
 * Este contrato corre nativamente en los nodos físicos de la blockchain Oxy•gen
 */
contract OXGToken is ERC20, Ownable {

    // Dirección del GreenPool (fondo ambiental)
    address public greenPoolWallet;

    // Porcentaje de fee / burn hacia GreenPool (5%)
    uint256 public feePercent = 5;

    // Direcciones exentas de fee (ej: staking, DAO, exchanges)
    mapping(address => bool) public isFeeExempt;

    // Dirección del contrato de staking para recompensas a validadores
    address public stakingContract;

    event FeeExemptionUpdated(address indexed account, bool exempt);
    event GreenPoolWalletUpdated(address indexed newWallet);
    event StakingContractUpdated(address indexed newContract);

    constructor(
        address _greenPoolWallet,
        uint256 initialSupply
    ) ERC20("Oxy-gen Economy Token", "OXG") {
        require(_greenPoolWallet != address(0), "GreenPool wallet required");
        greenPoolWallet = _greenPoolWallet;

        // Mint inicial al deployer
        _mint(msg.sender, initialSupply);
    }

    /**
     * @dev Permite al owner cambiar el GreenPool wallet
     * @notice En producción, esto debería ser controlado por la DAO
     */
    function setGreenPoolWallet(address _wallet) external onlyOwner {
        require(_wallet != address(0), "Invalid address");
        greenPoolWallet = _wallet;
        emit GreenPoolWalletUpdated(_wallet);
    }

    /**
     * @dev Permite establecer exenciones de fee para ciertas direcciones
     */
    function setFeeExemption(address account, bool exempt) external onlyOwner {
        isFeeExempt[account] = exempt;
        emit FeeExemptionUpdated(account, exempt);
    }

    /**
     * @dev Establecer contrato de staking para recompensas a validadores
     */
    function setStakingContract(address _stakingContract) external onlyOwner {
        require(_stakingContract != address(0), "Invalid address");
        stakingContract = _stakingContract;
        emit StakingContractUpdated(_stakingContract);
    }

    /**
     * @dev Aplicar fee interno en transferencias
     */
    function _applyFeeInternal(address from, uint256 amount) internal {
        uint256 fee = (amount * feePercent) / 100;
        if (fee > 0) {
            // Ajuste directo de balances internos
            _balances[from] -= fee;
            _balances[greenPoolWallet] += fee;

            // Disparar evento Transfer para compatibilidad ERC-20
            emit Transfer(from, greenPoolWallet, fee);
        }
    }

    /**
     * @dev Override del hook para aplicar fee
     */
    function _beforeTokenTransfer(
        address from,
        address to,
        uint256 amount
    ) internal override {
        // Aplicar fee solo si:
        // - No es mint
        // - No es burn
        // - El remitente no está exento
        // - No es el GreenPool mismo
        // - No es staking (para evitar doble fee)
        if (from != address(0) && 
            to != address(0) && 
            !isFeeExempt[from] && 
            from != greenPoolWallet &&
            from != stakingContract &&
            to != stakingContract) {
            _applyFeeInternal(from, amount);
        }

        super._beforeTokenTransfer(from, to, amount);
    }

    /**
     * @dev Función para distribuir recompensas a validadores (llamada por staking contract)
     */
    function distributeValidatorRewards(address[] calldata validators, uint256[] calldata amounts) external {
        require(msg.sender == stakingContract, "Only staking contract can distribute rewards");
        require(validators.length == amounts.length, "Arrays length mismatch");

        for (uint256 i = 0; i < validators.length; i++) {
            if (validators[i] != address(0) && amounts[i] > 0) {
                _mint(validators[i], amounts[i]);
            }
        }
    }
}

