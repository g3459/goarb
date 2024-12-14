import {IPoolFactory as IVeloV2Factory} from "./interfaces/velodrome-finance/contracts/contracts/interfaces/factories/IPoolFactory.sol";
import {IPool as IVeloV2Pool} from "./interfaces/velodrome-finance/contracts/contracts/interfaces/IPool.sol";
import {ICLFactory as IVeloV3Factory} from "./interfaces/velodrome-finance/slipstream/contracts/core/interfaces/ICLFactory.sol";
import {ICLPool as IVeloV3Pool} from "./interfaces/velodrome-finance/slipstream/contracts/core/interfaces/ICLPool.sol";
import "./interfaces/Uniswap/v2-core/contracts/interfaces/IUniswapV2Pair.sol";
import "./interfaces/Uniswap/v2-core/contracts/interfaces/IUniswapV2Factory.sol";
import "./interfaces/Uniswap/v3-core/contracts/interfaces/IUniswapV3Pool.sol";
import "./interfaces/Uniswap/v3-core/contracts/interfaces/IUniswapV3Factory.sol";
import "./interfaces/cryptoalgebra/Algebra/src/core/contracts/interfaces/IAlgebraFactory.sol";
import "./interfaces/cryptoalgebra/Algebra/src/core/contracts/interfaces/IAlgebraPool.sol";
import "./interfaces/openzeppelin/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";

contract CCaller {
    uint256 internal constant STATE_MASK = 0x7fffffff00000000000000000000000000000000000000000000000000000000;
    uint256 internal constant DIREC_MASK = 0x8000000000000000000000000000000000000000000000000000000000000000;
    uint256 internal constant PID_MASK = 0xff000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV2_PID = 0x01000000000000000000000000000000000000000000000000000000;
    uint256 internal constant UNIV3_PID = 0;
    uint256 internal constant ALGB_PID = 0x02000000000000000000000000000000000000000000000000000000;
    uint256 internal constant TRANSFER_SEL = 0xa9059cbb00000000000000000000000000000000000000000000000000000000;

    address internal immutable owner;

    constructor() payable {
        unchecked {
            owner = msg.sender;
        }
    }

    // function test(bytes32 a)public view returns(int){
    //     return int(uint(a));
    // }

    fallback() external payable check {
        unchecked {
            for (uint256 i; i <= msg.data.length - 0x20; i += 0x20) {
                uint256 poolCall;
                assembly {
                    poolCall := calldataload(i)
                }
                uint256 pid = poolCall & PID_MASK;
                bytes4 sel = pid == UNIV2_PID ? IUniswapV2Pair.getReserves.selector : pid == UNIV3_PID ? IUniswapV3Pool(address(0)).slot0.selector : IAlgebraPool(address(0)).globalState.selector;
                assembly {
                    mstore(0x80, sel)
                    pop(staticcall(gas(), poolCall, 0x80, 0x04, 0x80, 0x20))
                    if xor(and(keccak256(0x80, 0x20), STATE_MASK), and(poolCall, STATE_MASK)) {
                        revert(0, 0)
                    }
                }
            }
            // assembly{mstore(0x40,0x80)}
            uint256 amIn;
            assembly {
                amIn := and(calldataload(sub(calldatasize(), 0x20)), sub(shl(128, 1), 1))
            }
            for (uint256 i; i <= msg.data.length - 0x20; i += 0x20) {
                uint256 poolCall;
                assembly {
                    poolCall := calldataload(i)
                }
                address pool = address(uint160(poolCall));
                bool direc;
                assembly {
                    direc := and(poolCall, DIREC_MASK)
                }
                if (poolCall & PID_MASK == UNIV2_PID) {
                    bytes4 tokenSel = direc ? IUniswapV2Pair.token0.selector : IUniswapV2Pair.token1.selector;
                    assembly {
                        mstore(0x80, tokenSel)
                        pop(staticcall(gas(), pool, 0x80, 0x04, 0x80, 0x20))
                        let token := mload(0x80)
                        mstore(0x80, TRANSFER_SEL)
                        mstore(0x84, pool)
                        mstore(0xa4, amIn)
                        pop(call(gas(), token, 0, 0x80, 0x44, 0, 0))
                    }
                    bytes4 swapSel = IUniswapV2Pair.swap.selector;
                    assembly {
                        mstore(0x80, swapSel)
                        mstore(0xc4, address())
                        mstore(0xe4, 0x80)
                        mstore(0x104, 0)
                    }
                    amIn = uint256(uint48(poolCall >> 168)) << uint8(poolCall >> 160);
                    (uint256 amOut0, uint256 amOut1) = direc ? (uint256(0), amIn) : (amIn, uint256(0));
                    assembly {
                        mstore(0x84, amOut0)
                        mstore(0xa4, amOut1)
                        if iszero(call(gas(), pool, 0, 0x80, 0xa4, 0, 0)) {
                            revert(0, 0)
                        }
                    }
                } else {
                    bytes4 swapSel = IUniswapV3Pool(pool).swap.selector;
                    assembly {
                        mstore(0x80, swapSel)
                        mstore(0x84, address())
                        mstore(0xa4, direc)
                        mstore(0xc4, amIn)
                    }
                    uint256 sqrtL = direc ? 4295128740 : 1461446703485210103287273052203988822378723970341;
                    assembly {
                        mstore(0xe4, sqrtL)
                        mstore(0x104, 0xa0)
                        mstore(0x124, 0)
                        if iszero(call(gas(), pool, 0, 0x80, 0xc4, 0x80, 0x40)) {
                            revert(0, 0)
                        }
                    }
                    int256 am0;
                    int256 am1;
                    assembly {
                        am0 := mload(0x80)
                        am1 := mload(0xa0)
                    }
                    amIn = uint256(-(am0 < am1 ? am0 : am1));
                    require(amIn >= uint256(uint48(poolCall >> 168)) << uint8(poolCall >> 160));
                }
            }
            counter++;
        }
    }

    // receive() external payable{}

    // function recover() external payable check{
    //     unchecked{payable(msg.sender).transfer(address(this).balance);}
    // }

    function execute(address target, bytes calldata call) external payable check returns (bool s) {
        unchecked {
            (s, ) = target.call(call);
        }
    }

    modifier check() {
        _;
        unchecked {
            require(owner == tx.origin);
        }
    }

    modifier swapCallback(int256 am0, int256 am1) {
        _;
        unchecked {
            (bytes4 tokenSel, uint256 amIn) = am0 > am1 ? (IUniswapV2Pair.token0.selector, uint256(am0)) : (IUniswapV2Pair.token1.selector, uint256(am1));
            assembly {
                mstore(0x80, tokenSel)
                pop(call(gas(), caller(), 0, 0x80, 0x04, 0x80, 0x20))
                let token := mload(0x80)
                mstore(0x80, TRANSFER_SEL)
                mstore(0x84, caller())
                mstore(0xa4, amIn)
                pop(call(gas(), token, 0, 0x80, 0x44, 0, 0))
            }
        }
    }

    function uniswapV3SwapCallback(
        int256 am0,
        int256 am1,
        bytes calldata
    ) external payable swapCallback(am0, am1) check {}

    function pancakeV3SwapCallback(
        int256 am0,
        int256 am1,
        bytes calldata
    ) external payable swapCallback(am0, am1) check {}

    function algebraSwapCallback(
        int256 am0,
        int256 am1,
        bytes calldata
    ) external payable swapCallback(am0, am1) check {}
}
