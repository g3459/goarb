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

    uint internal constant STATE_MASK=0x7fffffff00000000000000000000000000000000000000000000000000000000;
    uint internal constant PID_MASK=0xff000000000000000000000000000000000000000000000000000000;
    uint internal constant DIREC_MASK=0x8000000000000000000000000000000000000000000000000000000000000000;
    uint internal constant UNIV2_PID=0x01000000000000000000000000000000000000000000000000000000;
    uint internal constant UNIV3_PID=0;
    uint internal constant ALGB_PID=0x02000000000000000000000000000000000000000000000000000000;
    uint internal constant TRANSFER_SEL=0xa9059cbb00000000000000000000000000000000000000000000000000000000;

    address internal immutable owner;
    
    constructor() payable{
        unchecked{owner=msg.sender;}
    }

    fallback() external payable check{
        unchecked{
            uint fmp=0x80;
            for(uint i;i<msg.data.length;i+=32){
                uint poolCall;
                assembly{
                    poolCall:=calldataload(i)
                }
                uint pid = poolCall&PID_MASK;
                uint outsize;
                bytes4 sel;
                if(pid==UNIV2_PID){
                    outsize=0x40;
                    sel=IUniswapV2Pair(address(0)).getReserves.selector;
                }else{
                    outsize=0x20;
                    sel=pid==ALGB_PID?IAlgebraPool(address(0)).globalState.selector:IUniswapV3Pool(address(0)).slot0.selector;
                }
                assembly{
                    mstore(fmp,sel)
                    pop(staticcall(gas(), poolCall, fmp, 0x04, fmp, outsize))
                    if xor(and(keccak256(fmp,outsize),STATE_MASK),and(poolCall,STATE_MASK)){
                        revert(0,0)
                    }
                }
                fmp+=outsize;
            }
            uint fmp2=0x80;
            for(uint i;i<msg.data.length;i+=32){
                uint poolCall;
                assembly{
                    poolCall:=calldataload(i)
                }
                uint amIn=uint(uint48(poolCall>>168))<<uint8(poolCall>>160);
                address pool=address(uint160(poolCall));
                bool direc;
                assembly{direc:=and(poolCall,DIREC_MASK)}
                if(poolCall&PID_MASK==UNIV2_PID){
                    uint rIn;uint rOut;
                    assembly{
                        rIn:=mload(fmp2)
                        rOut:=mload(add(fmp2,0x20))
                    }
                    fmp2+=0x40;
                    bytes4 tokenSel=direc?IUniswapV2Pair.token0.selector:IUniswapV2Pair.token1.selector;
                    assembly{
                        mstore(fmp,tokenSel)
                        pop(staticcall(gas(), pool, fmp, 0x04, fmp, 0x20))
                        let token:=mload(fmp)
                        mstore(fmp,TRANSFER_SEL)
                        mstore(add(fmp,0x04),pool)
                        mstore(add(fmp,0x24),amIn)
                        pop(call(gas(), token, 0, fmp, 0x44, 0, 0))
                    }
                    uint amOut=(amIn-1)*997;
                    amOut = (amOut * rOut) / (rIn * 1000 + amOut) - 1;
                    bytes4 swapSel=IUniswapV2Pair.swap.selector;
                    assembly{
                        mstore(fmp,swapSel)
                        mstore(add(fmp,0x44),address())
                    }
                    if(direc){
                        assembly{mstore(add(fmp,0x24),amOut)}
                    }else{
                        assembly{mstore(add(fmp,0x04),amOut)}
                    }
                    assembly{
                        let s:=call(gas(), pool, 0, fmp, 0x64, 0, 0)
                        if iszero(s){
                            revert(0,0)
                        }
                    }
                }else{
                    fmp2+=0x20;
                    bytes4 swapSel=IUniswapV3Pool(pool).swap.selector;
                    assembly{
                        mstore(fmp,swapSel)
                        mstore(add(fmp,0x04),address())
                    }
                    uint sqrtL;
                    if(direc){
                        sqrtL=4295128740;
                        assembly{mstore(add(fmp,0x24),direc)}
                    }else{
                        sqrtL=1461446703485210103287273052203988822378723970341;
                    }
                    assembly{
                        mstore(add(fmp,0x44),amIn)
                        mstore(add(fmp,0x64),sqrtL)
                        let s:=call(gas(), pool, 0, fmp, 0x84, 0, 0)
                        if iszero(s){
                            revert(0,0)
                        }
                    }
                }
            }
        }
    }
    // receive() external payable{}

    // function recover() external payable check{
    //     unchecked{payable(msg.sender).transfer(address(this).balance);}
    // }

    function execute(address target, bytes calldata call) external payable check returns (bool s){
        unchecked{(s,)=target.call(call);}
    }

    modifier check{
        _;
        unchecked{require(owner==tx.origin);}
    }

    modifier swapCallback(int am0,int am1){
        _;
        unchecked{
            (bytes4 tokenSel,uint amIn)=am0>am1?(IUniswapV2Pair.token0.selector,uint(am0)):(IUniswapV2Pair.token1.selector,uint(am1));
            assembly{
                mstore(0x80,tokenSel)
                pop(call(gas(), caller(), 0, 0x80, 0x04, 0x80, 0x20))
                let token:=mload(0x80)
                mstore(0x80,TRANSFER_SEL)
                mstore(0x84,caller())
                mstore(0xa4,amIn)
                pop(call(gas(), token, 0, 0x80, 0x44, 0, 0))
            }    
        }
    }

    function uniswapV3SwapCallback(int am0 , int am1, bytes calldata) external payable swapCallback(am0,am1) check{}

    function pancakeV3SwapCallback(int am0 , int am1, bytes calldata) external payable swapCallback(am0,am1) check{}

    function algebraSwapCallback(int am0, int am1, bytes calldata) external payable swapCallback(am0,am1) check{}

}
