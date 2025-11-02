// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";
import "./OXG.sol";

/**
 * @title OxyDAO
 * @dev DAO (Organización Autónoma Descentralizada) de Oxy•gen Economy
 * 
 * Permite a los holders de OXG participar en gobernanza mediante staking y votación
 */
contract OxyDAO is Ownable {
    
    OXGToken public oxgToken;
    
    // Proposiciones de la DAO
    struct Proposal {
        uint256 proposalId;
        address proposer;
        string title;
        string description;
        uint256 votingStart;
        uint256 votingEnd;
        uint256 votesFor;
        uint256 votesAgainst;
        bool executed;
        ProposalType proposalType;
        bytes data; // Datos adicionales según el tipo de propuesta
    }

    enum ProposalType {
        GreenPoolProject,    // Aprobar proyecto en GreenPool
        ParameterChange,     // Cambiar parámetros del sistema
        TreasurySpending,    // Gastos de tesorería
        ContractUpgrade      // Actualizar contratos
    }

    mapping(uint256 => Proposal) public proposals;
    mapping(uint256 => mapping(address => bool)) public hasVoted;
    mapping(address => uint256) public stakedAmount; // Stake de OXG para votar
    
    uint256 public proposalCount;
    uint256 public minimumStake = 1000 * 10**18; // Mínimo 1000 OXG para votar
    uint256 public votingPeriod = 7 days; // Periodo de votación
    
    // Eventos
    event ProposalCreated(uint256 indexed proposalId, address indexed proposer, string title);
    event VoteCasted(uint256 indexed proposalId, address indexed voter, bool support, uint256 votes);
    event ProposalExecuted(uint256 indexed proposalId);
    event StakeDeposited(address indexed staker, uint256 amount);
    event StakeWithdrawn(address indexed staker, uint256 amount);
    
    constructor(address _oxgToken) {
        require(_oxgToken != address(0), "OXG token address required");
        oxgToken = OXGToken(_oxgToken);
    }

    /**
     * @dev Depositar stake para participar en votaciones
     */
    function depositStake(uint256 _amount) external {
        require(_amount > 0, "Amount must be greater than 0");
        
        // Transferir tokens al contrato (staking)
        oxgToken.transferFrom(msg.sender, address(this), _amount);
        
        stakedAmount[msg.sender] += _amount;
        emit StakeDeposited(msg.sender, _amount);
    }

    /**
     * @dev Retirar stake (después de cooldown)
     */
    function withdrawStake(uint256 _amount) external {
        require(stakedAmount[msg.sender] >= _amount, "Insufficient staked amount");
        
        stakedAmount[msg.sender] -= _amount;
        oxgToken.transfer(msg.sender, _amount);
        
        emit StakeWithdrawn(msg.sender, _amount);
    }

    /**
     * @dev Crear una nueva propuesta
     */
    function createProposal(
        string memory _title,
        string memory _description,
        ProposalType _proposalType,
        bytes memory _data
    ) external returns (uint256) {
        require(stakedAmount[msg.sender] >= minimumStake, "Insufficient stake to create proposal");
        
        proposalCount++;
        proposals[proposalCount] = Proposal({
            proposalId: proposalCount,
            proposer: msg.sender,
            title: _title,
            description: _description,
            votingStart: block.timestamp,
            votingEnd: block.timestamp + votingPeriod,
            votesFor: 0,
            votesAgainst: 0,
            executed: false,
            proposalType: _proposalType,
            data: _data
        });

        emit ProposalCreated(proposalCount, msg.sender, _title);
        return proposalCount;
    }

    /**
     * @dev Votar en una propuesta
     */
    function vote(uint256 _proposalId, bool _support) external {
        Proposal storage proposal = proposals[_proposalId];
        
        require(block.timestamp >= proposal.votingStart, "Voting has not started");
        require(block.timestamp <= proposal.votingEnd, "Voting has ended");
        require(!proposal.executed, "Proposal already executed");
        require(!hasVoted[_proposalId][msg.sender], "Already voted");
        require(stakedAmount[msg.sender] >= minimumStake, "Insufficient stake to vote");

        uint256 votingPower = stakedAmount[msg.sender];
        
        if (_support) {
            proposal.votesFor += votingPower;
        } else {
            proposal.votesAgainst += votingPower;
        }
        
        hasVoted[_proposalId][msg.sender] = true;
        
        emit VoteCasted(_proposalId, msg.sender, _support, votingPower);
    }

    /**
     * @dev Ejecutar propuesta aprobada
     */
    function executeProposal(uint256 _proposalId) external {
        Proposal storage proposal = proposals[_proposalId];
        
        require(block.timestamp > proposal.votingEnd, "Voting period has not ended");
        require(!proposal.executed, "Proposal already executed");
        require(proposal.votesFor > proposal.votesAgainst, "Proposal was not approved");

        proposal.executed = true;

        // Ejecutar acción según tipo de propuesta
        if (proposal.proposalType == ProposalType.GreenPoolProject) {
            // TODO: Interactuar con GreenPool para aprobar proyecto
        } else if (proposal.proposalType == ProposalType.ParameterChange) {
            // TODO: Cambiar parámetros
        } else if (proposal.proposalType == ProposalType.TreasurySpending) {
            // TODO: Ejecutar gasto de tesorería
        } else if (proposal.proposalType == ProposalType.ContractUpgrade) {
            // TODO: Actualizar contrato
        }

        emit ProposalExecuted(_proposalId);
    }

    /**
     * @dev Obtener información de una propuesta
     */
    function getProposal(uint256 _proposalId) external view returns (Proposal memory) {
        return proposals[_proposalId];
    }

    /**
     * @dev Cambiar stake mínimo requerido (solo owner)
     */
    function setMinimumStake(uint256 _minimumStake) external onlyOwner {
        minimumStake = _minimumStake;
    }

    /**
     * @dev Cambiar periodo de votación (solo owner)
     */
    function setVotingPeriod(uint256 _votingPeriod) external onlyOwner {
        votingPeriod = _votingPeriod;
    }

    /**
     * @dev Verificar si una dirección puede votar
     */
    function canVote(address _voter) external view returns (bool) {
        return stakedAmount[_voter] >= minimumStake;
    }

    /**
     * @dev Obtener voting power de una dirección
     */
    function getVotingPower(address _voter) external view returns (uint256) {
        return stakedAmount[_voter];
    }
}

